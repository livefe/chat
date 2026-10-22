package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"

	"gorm.io/gorm"
)

const supportLongWaitSec int64 = 120

func (s *SupportService) CountQueue(ctx context.Context) (int64, error) {
	return s.supportStateRepo.CountQueue(ctx)
}

func (s *SupportService) AdminListQueue(ctx context.Context, req *types.SupportQueueQuery) (*types.SupportAdminSessionListResponse, error) {
	rows, total, err := s.supportStateRepo.ListQueue(ctx, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, 0, rows, total), nil
}

func (s *SupportService) AdminListMine(ctx context.Context, employeeID uint, page, pageSize int) (*types.SupportAdminSessionListResponse, error) {
	rows, total, err := s.supportStateRepo.ListByEmployee(ctx, employeeID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total), nil
}

func (s *SupportService) AdminListMyClosed(ctx context.Context, employeeID uint, page, pageSize int) (*types.SupportAdminSessionListResponse, error) {
	rows, total, err := s.supportStateRepo.ListClosedByEmployee(ctx, employeeID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total), nil
}

func overviewFilter(req *types.SupportOverviewQuery) (repository.SupportOverviewFilter, error) {
	f := repository.SupportOverviewFilter{
		AssigneeEmployeeID: req.AssigneeEmployeeID,
		Keyword:            strings.TrimSpace(req.Keyword),
	}
	if req.StartDate != "" {
		t, err := time.ParseInLocation("2006-01-02", req.StartDate, time.Local)
		if err != nil {
			return f, errs.Invalid("开始日期格式不正确")
		}
		f.From = t.Unix()
	}
	if req.EndDate != "" {
		t, err := time.ParseInLocation("2006-01-02", req.EndDate, time.Local)
		if err != nil {
			return f, errs.Invalid("结束日期格式不正确")
		}
		f.To = t.AddDate(0, 0, 1).Unix()
	}
	if f.From > 0 && f.To > 0 && f.From >= f.To {
		return f, errs.Invalid("开始日期不能晚于结束日期")
	}
	return f, nil
}

func (s *SupportService) AdminListActive(ctx context.Context, employeeID uint, req *types.SupportOverviewQuery) (*types.SupportAdminSessionListResponse, error) {
	if err := s.requireViewAll(ctx, employeeID); err != nil {
		return nil, err
	}
	f, err := overviewFilter(req)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.supportStateRepo.ListActive(ctx, req.Page, req.PageSize, f)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total), nil
}

func (s *SupportService) AdminListClosed(ctx context.Context, employeeID uint, req *types.SupportOverviewQuery) (*types.SupportAdminSessionListResponse, error) {
	if err := s.requireViewAll(ctx, employeeID); err != nil {
		return nil, err
	}
	f, err := overviewFilter(req)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.supportStateRepo.ListClosed(ctx, req.Page, req.PageSize, f)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toAdminList(ctx, employeeID, rows, total), nil
}

func (s *SupportService) AdminGetSession(ctx context.Context, employeeID, conversationID uint) (*types.SupportAdminSessionItem, error) {
	if _, _, err := s.assertAgentAccess(ctx, employeeID, conversationID); err != nil {
		return nil, err
	}
	return s.adminSessionItem(ctx, employeeID, conversationID)
}

func (s *SupportService) AdminListByUser(ctx context.Context, employeeID uint, req *types.SupportByUserQuery) (*types.SupportUserSessionListResponse, error) {
	if err := s.requireViewAll(ctx, employeeID); err != nil {
		return nil, err
	}
	rows, total, err := s.supportStateRepo.ListByUser(ctx, req.UserID, 0, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toUserSessionList(ctx, rows, total), nil
}

// AdminListHistory 返回指定会话所属用户的其他客服会话，供坐席了解来访历史。
func (s *SupportService) AdminListHistory(ctx context.Context, employeeID uint, req *types.SupportHistoryQuery) (*types.SupportUserSessionListResponse, error) {
	st, _, err := s.assertAgentAccess(ctx, employeeID, req.ConversationID)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.supportStateRepo.ListByUser(ctx, st.UserID, req.ConversationID, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toUserSessionList(ctx, rows, total), nil
}

// AdminListAgents 返回可接待客服的员工（角色含接入权限或超级管理员），并标注是否在线。
func (s *SupportService) AdminListAgents(ctx context.Context) (*types.SupportAgentListResponse, error) {
	ids, err := s.supportStateRepo.ListAgentEmployeeIDs(ctx, permCodeSupportAccept, types.SuperAdminRoleCode)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	emps, err := s.employeeRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	depIDs := make([]uint, 0, len(emps))
	for _, e := range emps {
		depIDs = append(depIDs, e.DepartmentID)
	}
	depNames := s.imSvc.departmentNames(ctx, depIDs)
	online := s.imSvc.OnlineEmployees(ctx, ids)
	list := make([]types.SupportAgentItem, 0, len(emps))
	for _, e := range emps {
		_, on := online[e.ID]
		list = append(list, types.SupportAgentItem{
			ID:             e.ID,
			Name:           e.Name,
			AvatarURL:      e.AvatarURL,
			DepartmentName: depNames[e.DepartmentID],
			Online:         on,
		})
	}
	return &types.SupportAgentListResponse{List: list}, nil
}

func (s *SupportService) AdminOverviewSummary(ctx context.Context, employeeID uint) (*types.SupportOverviewSummaryResponse, error) {
	if err := s.requireViewAll(ctx, employeeID); err != nil {
		return nil, err
	}
	queueCount, err := s.supportStateRepo.CountQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	activeCount, err := s.supportStateRepo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	longWait, err := s.supportStateRepo.CountQueueLongWait(ctx, supportLongWaitSec, time.Now().Unix())
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.SupportOverviewSummaryResponse{
		QueueCount:         queueCount,
		ActiveCount:        activeCount,
		LongWaitQueueCount: longWait,
	}, nil
}

func (s *SupportService) employeeName(ctx context.Context, id uint, fallback string) string {
	if id == 0 {
		return fallback
	}
	if e, err := s.employeeRepo.GetByID(ctx, id); err == nil && e != nil && e.Name != "" {
		return e.Name
	}
	return fallback
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
	name := s.employeeName(ctx, employeeID, "客服")
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

	return s.adminSessionItem(ctx, employeeID, conversationID)
}

func (s *SupportService) AdminTransfer(ctx context.Context, fromEmployeeID uint, req *types.SupportTransferRequest) error {
	st, err := s.supportStateRepo.GetByConversationID(ctx, req.ConversationID)
	if err != nil {
		return errs.NotFound("会话不存在")
	}
	conv, err := s.imSvc.GetConversation(ctx, req.ConversationID)
	if err != nil || conv.Status != model.IMConvStatusActive {
		return errs.Invalid("会话不可转接")
	}
	isAssignee := st.AssigneeEmployeeID == fromEmployeeID
	if !isAssignee && !s.hasViewAll(ctx, fromEmployeeID) {
		return errs.Forbidden("仅当前坐席或督导可转接")
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
		ok, err := s.permSvc.HasPermissionCode(ctx, req.ToEmployeeID, permCodeSupportAccept)
		if err != nil || !ok {
			return errs.Invalid("目标同事没有客服接入权限")
		}
	}
	now := time.Now().Unix()
	prevAssignee := st.AssigneeEmployeeID
	operatorName := s.employeeName(ctx, fromEmployeeID, "客服")
	toName := to.Name
	if toName == "" {
		toName = "客服"
	}
	sysText := fmt.Sprintf("%s 已将会话转接给 %s", operatorName, toName)
	if req.ToEmployeeID == fromEmployeeID {
		sysText = fmt.Sprintf("%s 已接管本会话", operatorName)
	} else if !isAssignee && prevAssignee > 0 {
		sysText = fmt.Sprintf("%s 已将会话从 %s 转接给 %s", operatorName, s.employeeName(ctx, prevAssignee, "原坐席"), toName)
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
	if prevAssignee > 0 && prevAssignee != req.ToEmployeeID {
		_ = s.imSvc.ClearUnread(ctx, prevAssignee, req.ConversationID)
	}
	s.notifyUserSession(ctx, st.UserID, req.ConversationID)
	return nil
}

func (s *SupportService) AdminClose(ctx context.Context, employeeID uint, req *types.SupportCloseRequest) error {
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
			"close_reason": req.Reason,
			"close_note":   strings.TrimSpace(req.Note),
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
	_ = s.imSvc.ClearUnread(ctx, employeeID, req.ConversationID)
	s.broadcastQueueUpdate(ctx)
	s.notifyUserSession(ctx, st.UserID, req.ConversationID)
	_ = s.clearUserUnread(ctx, st.UserID)
	return nil
}

func (s *SupportService) AdminListMessages(ctx context.Context, employeeID, conversationID, beforeID uint, limit int) (*types.IMMessageListResponse, error) {
	if _, _, err := s.assertAgentAccess(ctx, employeeID, conversationID); err != nil {
		return nil, err
	}
	return s.listMessages(ctx, conversationID, beforeID, limit)
}

func (s *SupportService) AdminSend(ctx context.Context, employeeID uint, req *types.SupportSendRequest) (*types.IMMessageItem, error) {
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

// AdminRead 只处理当前坐席自己的未读；督导旁观他人会话不产生已读记录。
func (s *SupportService) AdminRead(ctx context.Context, employeeID uint, req *types.SupportReadRequest) error {
	st, err := s.supportStateRepo.GetByConversationID(ctx, req.ConversationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("会话不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if st.AssigneeEmployeeID != employeeID {
		return errs.Forbidden("仅当前坐席可标记已读")
	}
	if err := s.imSvc.MarkRead(ctx, req.ConversationID, model.IMPrincipalEmployee, employeeID, req.LastReadMsgID); err != nil {
		return err
	}
	_ = s.imSvc.ClearUnread(ctx, employeeID, req.ConversationID)
	return nil
}
