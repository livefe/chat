package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (s *ReportService) AdminList(ctx context.Context, req *types.AdminReportListRequest) (*types.ReportListResponse, error) {
	list, total, err := s.reportRepo.List(ctx, repository.ReportListFilter{
		ReporterID:   req.ReporterID,
		TargetUserID: req.TargetUserID,
		TargetID:     req.TargetID,
		Status:       req.Status,
		TargetType:   req.TargetType,
		Reason:       req.Reason,
		Keyword:      req.Keyword,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		OldestFirst:  req.Status != nil && *req.Status == model.ReportStatusPending,
		Page:         req.Page,
		PageSize:     req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询举报列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询举报列表失败: %w", err)
	}
	items, err := s.enrichReportItems(ctx, list)
	if err != nil {
		return nil, err
	}
	return &types.ReportListResponse{Total: total, List: items}, nil
}

func (s *ReportService) AdminGet(ctx context.Context, id uint) (*types.ReportItem, error) {
	item, err := s.reportRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("举报不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.enrichReportItems(ctx, []*model.Report{item})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	result := items[0]
	if item.TargetUserID > 0 {
		summary, err := s.targetUserSummary(ctx, item.TargetUserID)
		if err != nil {
			return nil, err
		}
		result.TargetUser = summary
	}
	return &result, nil
}

func (s *ReportService) targetUserSummary(ctx context.Context, userID uint) (*types.UserGovernanceSummary, error) {
	summary, err := s.userSvc.GovernanceSummary(ctx, userID)
	if err != nil {
		if e, ok := errs.As(err); ok && e.Code == errs.CodeNotFound {
			return nil, nil
		}
		return nil, err
	}
	count, err := s.reportRepo.CountByTargetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	summary.ReportCount = count
	return summary, nil
}

func (s *ReportService) AdminHandle(ctx context.Context, employeeID uint, req *types.AdminReportHandleRequest) (*types.AdminReportHandleResult, error) {
	item, err := s.reportRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("举报不存在")
		}
		return nil, fmt.Errorf("处理失败: %w", err)
	}
	if item.Status != model.ReportStatusPending {
		return nil, errs.Conflict("该举报已处理")
	}
	if req.Status != model.ReportStatusResolved && req.Status != model.ReportStatusRejected {
		return nil, errs.Invalid("处理状态无效")
	}

	handleResult := strings.TrimSpace(req.HandleResult)
	if utf8.RuneCountInString(handleResult) > reportContentMaxLen {
		return nil, errs.Invalid("处理说明过长")
	}

	action := req.Action
	if req.Status == model.ReportStatusRejected {
		action = model.ReportHandleActionNone
	}
	if req.Status == model.ReportStatusResolved {
		meta := PunishMeta{
			EmployeeID: employeeID,
			Reason:     reportPunishReason(item, handleResult),
			ReportID:   item.ID,
		}
		if err := s.applyHandleAction(ctx, item, action, req.MuteDuration, meta); err != nil {
			s.log.Error(ctx, "举报处置动作失败", logger.Uint("report_id", item.ID), logger.Error(err))
			return nil, err
		}
	}

	now := time.Now().Unix()
	item.Status = req.Status
	item.HandleResult = handleResult
	item.HandleAction = action
	item.EmployeeID = employeeID
	item.HandledAt = now
	if err := s.reportRepo.Update(ctx, item); err != nil {
		s.log.Error(ctx, "更新举报状态失败", logger.Error(err))
		return nil, fmt.Errorf("处理失败: %w", err)
	}

	result := &types.AdminReportHandleResult{}
	if req.ResolveSameTarget {
		n, err := s.reportRepo.ResolvePendingByTarget(ctx,
			repository.ReportTargetKey{TargetType: item.TargetType, TargetID: item.TargetID}, item.ID,
			map[string]any{
				"status":        req.Status,
				"handle_result": handleResult,
				"handle_action": action,
				"employee_id":   employeeID,
				"handled_at":    now,
			})
		if err != nil {
			s.log.Error(ctx, "同对象举报结案失败", logger.Uint("report_id", item.ID), logger.Error(err))
			return nil, fmt.Errorf("本条已处理，但同对象举报结案失败: %w", err)
		}
		result.ResolvedCount = n
	}
	return result, nil
}

func reportPunishReason(item *model.Report, handleResult string) string {
	if handleResult != "" {
		return handleResult
	}
	return fmt.Sprintf("经核实存在「%s」行为", utils.TextOf(reportReasonText, item.Reason))
}

func (s *ReportService) applyHandleAction(ctx context.Context, item *model.Report, action uint8, muteDuration int, meta PunishMeta) error {
	needUser := action == model.ReportHandleActionWarn ||
		action == model.ReportHandleActionMute ||
		action == model.ReportHandleActionDisableUser
	if needUser && item.TargetUserID == 0 {
		return errs.Invalid("缺少被举报用户")
	}
	switch action {
	case model.ReportHandleActionNone:
		return nil
	case model.ReportHandleActionWarn:
		return s.userSvc.AdminWarn(ctx, item.TargetUserID, meta)
	case model.ReportHandleActionDeleteContent:
		if item.TargetType != model.ReportTargetMessage {
			return errs.Invalid("仅消息举报可删除内容")
		}
		return s.messageSvc.DeleteForReport(ctx, item.TargetID, meta)
	case model.ReportHandleActionMute:
		if muteDuration == 0 {
			return errs.Invalid("请选择禁言时长")
		}
		return s.userSvc.AdminMute(ctx, item.TargetUserID, muteDuration, meta)
	case model.ReportHandleActionDisableUser:
		return s.userSvc.DisableForReport(ctx, item.TargetUserID, meta)
	case model.ReportHandleActionDisablePost:
		if item.TargetType != model.ReportTargetPost {
			return errs.Invalid("仅帖子举报可下架")
		}
		if err := s.postSvc.DisableByReport(ctx, item.TargetID); err != nil {
			return err
		}
		meta.TargetID = item.TargetID
		return s.userSvc.RecordContentPunishment(ctx, item.TargetUserID, model.PunishActionDisablePost, meta)
	default:
		return errs.Invalid("处置动作无效")
	}
}

func (s *ReportService) enrichReportItems(ctx context.Context, list []*model.Report) ([]types.ReportItem, error) {
	userIDs := make([]uint, 0, len(list)*2)
	employeeIDs := make([]uint, 0, len(list))
	targetKeys := make([]repository.ReportTargetKey, 0, len(list))
	for _, item := range list {
		userIDs = append(userIDs, item.ReporterID, item.TargetUserID)
		employeeIDs = append(employeeIDs, item.EmployeeID)
		targetKeys = append(targetKeys, repository.ReportTargetKey{TargetType: item.TargetType, TargetID: item.TargetID})
	}

	profiles, err := s.userSvc.GetProfiles(ctx, uniqueUints(userIDs))
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	employeeMap := make(map[uint]*model.Employee)
	if ids := uniqueUints(employeeIDs); len(ids) > 0 {
		employees, err := s.employeeRepo.GetByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		for _, e := range employees {
			employeeMap[e.ID] = e
		}
	}
	counts, err := s.reportRepo.CountByTargets(ctx, targetKeys)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	items := make([]types.ReportItem, 0, len(list))
	for _, item := range list {
		count := counts[repository.ReportTargetKey{TargetType: item.TargetType, TargetID: item.TargetID}]
		items = append(items, s.toReportItem(ctx, item, profiles, employeeMap[item.EmployeeID], count))
	}
	return items, nil
}

func (s *ReportService) toReportItem(ctx context.Context, item *model.Report, profiles map[uint]types.UserProfile, employee *model.Employee, targetCount int64) types.ReportItem {
	ids := parseMediaIDs(item.Images)
	images := []string{}
	if s.mediaSvc != nil {
		images = s.mediaSvc.ResolveURLList(ctx, ids)
	}
	result := types.ReportItem{
		ID:                item.ID,
		ReporterID:        item.ReporterID,
		TargetType:        item.TargetType,
		TargetTypeText:    utils.TextOf(reportTargetTypeText, item.TargetType),
		TargetID:          item.TargetID,
		TargetUserID:      item.TargetUserID,
		Reason:            item.Reason,
		ReasonText:        utils.TextOf(reportReasonText, item.Reason),
		Content:           item.Content,
		Images:            images,
		Context:           parseReportContext(item.Context),
		Status:            item.Status,
		StatusText:        utils.TextOf(reportStatusText, item.Status),
		HandleResult:      item.HandleResult,
		HandleAction:      item.HandleAction,
		HandleActionText:  utils.TextOf(reportHandleActionText, item.HandleAction),
		EmployeeID:        item.EmployeeID,
		TargetReportCount: targetCount,
		CreatedAt:         utils.FormatDateTime(item.CreatedAt),
		UpdatedAt:         utils.FormatDateTime(item.UpdatedAt),
	}
	if item.HandledAt > 0 {
		result.HandledAt = utils.FormatDateTime(time.Unix(item.HandledAt, 0))
	}
	if reporter, ok := profiles[item.ReporterID]; ok {
		result.ReporterNickname = utils.DisplayName(reporter.Nickname)
		result.ReporterMobile = utils.DesensitizeMobile(reporter.Mobile)
	} else {
		result.ReporterNickname = deactivatedDisplayName()
	}
	if item.TargetUserID > 0 {
		if target, ok := profiles[item.TargetUserID]; ok {
			result.TargetUserName = utils.DisplayName(target.Nickname)
			result.TargetUserMobile = utils.DesensitizeMobile(target.Mobile)
		} else {
			result.TargetUserName = deactivatedDisplayName()
		}
	}
	if employee != nil {
		result.EmployeeName = employee.Name
	}
	return result
}

func parseReportContext(raw datatypes.JSON) types.ReportContextSnapshot {
	if len(raw) == 0 {
		return types.ReportContextSnapshot{}
	}
	var snapshot types.ReportContextSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return types.ReportContextSnapshot{}
	}
	return snapshot
}
