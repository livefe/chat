package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
)

func (s *SupportService) CountQueue(ctx context.Context) (int64, error) {
	return s.supportStateRepo.CountQueue(ctx)
}

func (s *SupportService) AdminListQueue(ctx context.Context, page, pageSize int) (*types.SupportAdminSessionListResponse, error) {
	rows, total, err := s.supportStateRepo.ListQueue(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, 0, rows, total)
}

func (s *SupportService) AdminListMine(ctx context.Context, employeeID uint, page, pageSize int) (*types.SupportAdminSessionListResponse, error) {
	rows, total, err := s.supportStateRepo.ListByEmployee(ctx, employeeID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total)
}

func (s *SupportService) AdminListMyClosed(ctx context.Context, employeeID uint, page, pageSize int) (*types.SupportAdminSessionListResponse, error) {
	rows, total, err := s.supportStateRepo.ListClosedByEmployee(ctx, employeeID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total)
}

func (s *SupportService) AdminListActive(ctx context.Context, employeeID uint, page, pageSize int, assigneeEmployeeID uint, keyword string) (*types.SupportAdminSessionListResponse, error) {
	if ok, err := s.hasViewAll(ctx, employeeID); err != nil || !ok {
		return nil, errs.Forbidden("无权查看全部客服会话")
	}
	rows, total, err := s.supportStateRepo.ListActive(ctx, page, pageSize, repository.SupportOverviewFilter{
		AssigneeEmployeeID: assigneeEmployeeID,
		Keyword:            keyword,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total)
}

func (s *SupportService) AdminListClosed(ctx context.Context, employeeID uint, page, pageSize int, assigneeEmployeeID uint, keyword string) (*types.SupportAdminSessionListResponse, error) {
	if ok, err := s.hasViewAll(ctx, employeeID); err != nil || !ok {
		return nil, errs.Forbidden("无权查看全部客服会话")
	}
	rows, total, err := s.supportStateRepo.ListClosed(ctx, page, pageSize, repository.SupportOverviewFilter{
		AssigneeEmployeeID: assigneeEmployeeID,
		Keyword:            keyword,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total)
}

func (s *SupportService) AdminOverviewSummary(ctx context.Context, employeeID uint) (*types.SupportOverviewSummaryResponse, error) {
	if ok, err := s.hasViewAll(ctx, employeeID); err != nil || !ok {
		return nil, errs.Forbidden("无权查看客服总览")
	}
	const longWaitSec int64 = 120
	now := time.Now().Unix()
	queueCount, err := s.supportStateRepo.CountQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	activeCount, err := s.supportStateRepo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	longWait, err := s.supportStateRepo.CountQueueLongWait(ctx, longWaitSec, now)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.SupportOverviewSummaryResponse{
		QueueCount:         queueCount,
		ActiveCount:        activeCount,
		LongWaitQueueCount: longWait,
	}, nil
}

func (s *SupportService) AdminAccept(ctx context.Context, employeeID, conversationID uint) (*types.SupportAdminSessionItem, error) {
	st, err := s.supportStateRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, conversationID)
	if err != nil || conv.Type != model.IMConvTypeSupport {
		return nil, errs.NotFound("会话不存在")
	}
	if conv.Status != model.IMConvStatusQueued {
		return nil, errs.Conflict("会话已被接入或不在排队中")
	}
	now := time.Now().Unix()
	emp, _ := s.employeeRepo.GetByID(ctx, employeeID)
	name := "客服"
	if emp != nil && emp.Name != "" {
		name = emp.Name
	}
	var sys *types.IMMessageItem
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		ok, err := s.imSvc.SetConversationStatusIf(ctx, conversationID, model.IMConvStatusQueued, model.IMConvStatusActive)
		if err != nil {
			return fmt.Errorf("接入失败: %w", err)
		}
		if !ok {
			return errs.Conflict("会话已被接入或不在排队中")
		}
		if err := s.supportStateRepo.Update(ctx, conversationID, map[string]any{
			"assignee_employee_id": employeeID,
			"accept_at":            now,
		}); err != nil {
			return fmt.Errorf("接入失败: %w", err)
		}
		if err := s.imSvc.ReactivateMember(ctx, conversationID, model.IMPrincipalEmployee, employeeID, model.IMMemberRoleAgent, now); err != nil {
			return fmt.Errorf("接入失败: %w", err)
		}
		sys, err = s.writeSystemMessage(ctx, conversationID, fmt.Sprintf("%s 已接入为您服务", name))
		return err
	}); err != nil {
		return nil, err
	}
	if sys != nil {
		s.fanoutMessage(ctx, conversationID, model.IMPrincipalSystem, 0, sys)
	}

	s.broadcastQueueUpdate(ctx)
	s.notifyUserSession(ctx, st.UserID, conversationID)

	item, err := s.adminSessionItem(ctx, employeeID, conversationID)
	return item, err
}

func (s *SupportService) AdminTransfer(ctx context.Context, fromEmployeeID uint, req *types.SupportTransferRequest) error {
	if req == nil {
		return errs.Invalid("参数错误")
	}
	st, err := s.supportStateRepo.GetByConversationID(ctx, req.ConversationID)
	if err != nil {
		return errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, req.ConversationID)
	if err != nil || conv.Status != model.IMConvStatusActive {
		return errs.Invalid("会话不可转接")
	}
	isAssignee := st.AssigneeEmployeeID == fromEmployeeID
	if !isAssignee {
		ok, err := s.hasViewAll(ctx, fromEmployeeID)
		if err != nil || !ok {
			return errs.Forbidden("仅当前坐席或督导可转接")
		}
	}
	if req.ToEmployeeID == 0 {
		return errs.Invalid("转接对象无效")
	}

	if req.ToEmployeeID == fromEmployeeID {
		if isAssignee {
			return errs.Invalid("转接对象无效")
		}
	} else if req.ToEmployeeID == st.AssigneeEmployeeID {
		return errs.Invalid("目标已是当前坐席")
	}
	to, err := s.employeeRepo.GetByID(ctx, req.ToEmployeeID)
	if err != nil || to.Status != model.EmployeeStatusEnabled {
		return errs.Invalid("目标坐席不可用")
	}
	if s.permSvc != nil {
		ok, err := s.permSvc.CheckPermission(ctx, req.ToEmployeeID, "/api/admin/support/accept", "POST")
		if err != nil || !ok {
			return errs.Invalid("目标同事没有客服接入权限")
		}
	}
	now := time.Now().Unix()
	prevAssignee := st.AssigneeEmployeeID
	operatorName := "客服"
	if from, err := s.employeeRepo.GetByID(ctx, fromEmployeeID); err == nil && from != nil && from.Name != "" {
		operatorName = from.Name
	}
	toName := to.Name
	if toName == "" {
		toName = "客服"
	}
	sysText := fmt.Sprintf("%s 已将会话转接给 %s", operatorName, toName)
	if req.ToEmployeeID == fromEmployeeID {
		sysText = fmt.Sprintf("%s 已接管本会话", operatorName)
	} else if !isAssignee && prevAssignee > 0 {
		prevName := "原坐席"
		if e, err := s.employeeRepo.GetByID(ctx, prevAssignee); err == nil && e != nil && e.Name != "" {
			prevName = e.Name
		}
		sysText = fmt.Sprintf("%s 已将会话从 %s 转接给 %s", operatorName, prevName, toName)
	}
	var sys *types.IMMessageItem
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if prevAssignee > 0 && prevAssignee != req.ToEmployeeID {
			if err := s.imSvc.LeaveMember(ctx, req.ConversationID, model.IMPrincipalEmployee, prevAssignee, now); err != nil {
				return fmt.Errorf("转接失败: %w", err)
			}
		}
		if err := s.imSvc.ReactivateMember(ctx, req.ConversationID, model.IMPrincipalEmployee, req.ToEmployeeID, model.IMMemberRoleAgent, now); err != nil {
			return fmt.Errorf("转接失败: %w", err)
		}
		if err := s.supportStateRepo.Update(ctx, req.ConversationID, map[string]any{
			"assignee_employee_id": req.ToEmployeeID,
		}); err != nil {
			return fmt.Errorf("转接失败: %w", err)
		}
		var err error
		sys, err = s.writeSystemMessage(ctx, req.ConversationID, sysText)
		return err
	}); err != nil {
		return err
	}
	if sys != nil {
		s.fanoutMessage(ctx, req.ConversationID, model.IMPrincipalSystem, 0, sys)
	}
	s.notifyUserSession(ctx, st.UserID, req.ConversationID)
	return nil
}

func (s *SupportService) AdminClose(ctx context.Context, employeeID uint, req *types.SupportCloseRequest) error {
	if req == nil {
		return errs.Invalid("参数错误")
	}
	st, err := s.supportStateRepo.GetByConversationID(ctx, req.ConversationID)
	if err != nil {
		return errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, req.ConversationID)
	if err != nil {
		return err
	}
	if conv.Status == model.IMConvStatusClosed {
		return nil
	}
	if conv.Status != model.IMConvStatusActive {
		return errs.Invalid("仅进行中的会话可结案，排队中请等待用户取消或接入后处理")
	}
	if st.AssigneeEmployeeID != employeeID {
		return errs.Forbidden("仅当前坐席可结案")
	}
	now := time.Now().Unix()
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "agent_close"
	}
	var sys *types.IMMessageItem
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		ok, err := s.imSvc.SetConversationStatusIf(ctx, req.ConversationID, model.IMConvStatusActive, model.IMConvStatusClosed)
		if err != nil {
			return fmt.Errorf("结案失败: %w", err)
		}
		if !ok {
			return errs.Conflict("会话状态已变更，请刷新后重试")
		}
		if err := s.supportStateRepo.Update(ctx, req.ConversationID, map[string]any{
			"close_at":     now,
			"close_reason": reason,
		}); err != nil {
			return fmt.Errorf("结案失败: %w", err)
		}
		sys, err = s.writeSystemMessage(ctx, req.ConversationID, "本次客服会话已结束，如需帮助请再次联系客服")
		return err
	}); err != nil {
		return err
	}
	if sys != nil {
		s.fanoutMessage(ctx, req.ConversationID, model.IMPrincipalSystem, 0, sys)
	}
	s.broadcastQueueUpdate(ctx)
	s.notifyUserSession(ctx, st.UserID, req.ConversationID)
	_ = s.clearUserUnread(ctx, st.UserID)
	return nil
}

func (s *SupportService) AdminListMessages(ctx context.Context, employeeID, conversationID, beforeID uint, limit int) (*types.IMMessageListResponse, error) {
	if err := s.assertAgentAccess(ctx, employeeID, conversationID); err != nil {
		return nil, err
	}
	return s.listMessages(ctx, conversationID, beforeID, limit)
}

func (s *SupportService) AdminSend(ctx context.Context, employeeID uint, req *types.SupportSendRequest) (*types.IMMessageItem, error) {
	if req == nil {
		return nil, errs.Invalid("参数错误")
	}
	st, err := s.supportStateRepo.GetByConversationID(ctx, req.ConversationID)
	if err != nil {
		return nil, errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, req.ConversationID)
	if err != nil || conv.Status != model.IMConvStatusActive {
		return nil, errs.Invalid("会话不可回复")
	}
	if st.AssigneeEmployeeID != employeeID {
		return nil, errs.Forbidden("仅当前坐席可回复")
	}
	if err := s.imSvc.RequireMember(ctx, req.ConversationID, model.IMPrincipalEmployee, employeeID); err != nil {
		return nil, err
	}
	return s.sendAs(ctx, model.IMPrincipalEmployee, employeeID, req)
}

func (s *SupportService) AdminRead(ctx context.Context, employeeID uint, req *types.SupportReadRequest) error {
	if req == nil {
		return errs.Invalid("参数错误")
	}
	if err := s.assertAgentAccess(ctx, employeeID, req.ConversationID); err != nil {
		return err
	}
	if err := s.imSvc.MarkRead(ctx, req.ConversationID, model.IMPrincipalEmployee, employeeID, req.LastReadMsgID); err != nil {
		return err
	}
	_ = s.imSvc.ClearUnread(ctx, employeeID, req.ConversationID)
	return nil
}
