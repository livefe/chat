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
		Status:     req.Status,
		TargetType: req.TargetType, Reason: req.Reason, Keyword: req.Keyword,
		Page: req.Page, PageSize: req.PageSize,
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
	targetCount, _ := s.reportRepo.CountByTarget(ctx, item.TargetType, item.TargetID)
	items, err := s.enrichReportItems(ctx, []*model.Report{item})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("查询失败")
	}
	result := items[0]
	result.TargetReportCount = targetCount
	return &result, nil
}

func (s *ReportService) AdminHandle(ctx context.Context, employeeID uint, req *types.AdminReportHandleRequest) error {
	item, err := s.reportRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("举报不存在")
		}
		return fmt.Errorf("处理失败: %w", err)
	}
	if item.Status == model.ReportStatusResolved || item.Status == model.ReportStatusRejected {
		return errs.Conflict("该举报已处理")
	}
	if req.Status != model.ReportStatusResolved && req.Status != model.ReportStatusRejected {
		return errs.Invalid("处理状态无效")
	}

	handleResult := strings.TrimSpace(req.HandleResult)
	if utf8.RuneCountInString(handleResult) > reportContentMaxLen {
		return errs.Invalid("处理说明过长")
	}

	action := req.Action
	if req.Status == model.ReportStatusRejected {
		action = model.ReportHandleActionNone
	}

	if req.Status == model.ReportStatusResolved {
		if err := s.applyHandleAction(ctx, item, action, req.MuteHours); err != nil {
			s.log.Error(ctx, "举报处置动作失败", logger.Uint("report_id", item.ID), logger.Error(err))
			return err
		}
	}

	item.Status = req.Status
	item.HandleResult = handleResult
	item.HandleAction = action
	item.EmployeeID = employeeID
	item.HandledAt = time.Now().Unix()

	if err := s.reportRepo.Update(ctx, item); err != nil {
		s.log.Error(ctx, "更新举报状态失败", logger.Error(err))
		return fmt.Errorf("处理失败: %w", err)
	}
	return nil
}

func (s *ReportService) applyHandleAction(ctx context.Context, item *model.Report, action uint8, muteHours int) error {
	switch action {
	case model.ReportHandleActionNone, model.ReportHandleActionWarn:
		return nil
	case model.ReportHandleActionDeleteContent:
		if item.TargetType != model.ReportTargetMessage {
			return errs.Invalid("仅消息举报可删除内容")
		}
		return s.messageSvc.AdminDelete(ctx, item.TargetID)
	case model.ReportHandleActionMute:
		if item.TargetUserID == 0 {
			return errs.Invalid("缺少被举报用户")
		}
		if muteHours <= 0 {
			return errs.Invalid("请填写禁言时长")
		}
		return s.userSvc.AdminMute(ctx, item.TargetUserID, muteHours, "举报处理禁言")
	case model.ReportHandleActionDisableUser:
		if item.TargetUserID == 0 {
			return errs.Invalid("缺少被举报用户")
		}
		return s.userSvc.AdminUpdateStatus(ctx, item.TargetUserID, model.UserStatusDisabled)
	case model.ReportHandleActionDisablePost:
		if item.TargetType != model.ReportTargetPost {
			return errs.Invalid("仅动态举报可下架")
		}
		return s.postSvc.AdminUpdateStatus(ctx, item.TargetID, model.PostStatusDisabled)
	default:
		if action > model.ReportHandleActionDisablePost {
			return errs.Invalid("处置动作无效")
		}
		return nil
	}
}

func (s *ReportService) enrichReportItems(ctx context.Context, list []*model.Report) ([]types.ReportItem, error) {
	userIDs := make([]uint, 0, len(list)*2)
	employeeIDs := make([]uint, 0, len(list))
	userSeen := make(map[uint]struct{})
	employeeSeen := make(map[uint]struct{})

	for _, item := range list {
		for _, uid := range []uint{item.ReporterID, item.TargetUserID} {
			if uid == 0 {
				continue
			}
			if _, ok := userSeen[uid]; !ok {
				userSeen[uid] = struct{}{}
				userIDs = append(userIDs, uid)
			}
		}
		if item.EmployeeID > 0 {
			if _, ok := employeeSeen[item.EmployeeID]; !ok {
				employeeSeen[item.EmployeeID] = struct{}{}
				employeeIDs = append(employeeIDs, item.EmployeeID)
			}
		}
	}

	profiles, err := s.userSvc.GetProfiles(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	employeeMap := make(map[uint]*model.Employee)
	if len(employeeIDs) > 0 {
		employees, err := s.employeeRepo.GetByIDs(ctx, employeeIDs)
		if err == nil {
			for _, e := range employees {
				employeeMap[e.ID] = e
			}
		}
	}

	items := make([]types.ReportItem, 0, len(list))
	for _, item := range list {
		targetCount, _ := s.reportRepo.CountByTarget(ctx, item.TargetType, item.TargetID)
		items = append(items, s.toReportItem(ctx, item, profiles, employeeMap[item.EmployeeID], targetCount))
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
		result.ReporterNickname = reporter.Nickname
		result.ReporterMobile = reporter.Mobile
	}
	if target, ok := profiles[item.TargetUserID]; ok {
		result.TargetUserName = target.Nickname
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
