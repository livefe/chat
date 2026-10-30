package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var punishActionText = map[uint8]string{
	model.PunishActionWarn:          "警告",
	model.PunishActionMute:          "禁言",
	model.PunishActionUnmute:        "解除禁言",
	model.PunishActionDisable:       "禁用账号",
	model.PunishActionEnable:        "启用账号",
	model.PunishActionDeactivate:    "注销账号",
	model.PunishActionDeleteMessage: "删除消息",
	model.PunishActionDisablePost:   "下架帖子",
}

func (s *UserService) recordPunishment(ctx context.Context, userID uint, action uint8, muteUntil int64, meta PunishMeta) error {
	return s.punishRepo.Create(ctx, &model.UserPunishment{
		UserID:     userID,
		Action:     action,
		MuteUntil:  muteUntil,
		TargetID:   meta.TargetID,
		Reason:     truncateRunes(strings.TrimSpace(meta.Reason), 499),
		EmployeeID: meta.EmployeeID,
		ReportID:   meta.ReportID,
	})
}

func (s *UserService) RecordContentPunishment(ctx context.Context, userID uint, action uint8, meta PunishMeta) error {
	if userID == 0 {
		return nil
	}
	return s.recordPunishment(ctx, userID, action, 0, meta)
}

func (s *UserService) AdminListPunishments(ctx context.Context, req *types.AdminUserSubListRequest) (*types.UserPunishmentListResponse, error) {
	rows, total, err := s.punishRepo.ListByUserID(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		s.log.Error(ctx, "查询处罚记录失败", logger.Uint("user_id", req.UserID), logger.Error(err))
		return nil, fmt.Errorf("查询处罚记录失败: %w", err)
	}
	employeeIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		employeeIDs = append(employeeIDs, row.EmployeeID)
	}
	names := map[uint]string{}
	if ids := uniqueUints(employeeIDs); len(ids) > 0 {
		employees, err := s.employeeRepo.GetByIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("查询处罚记录失败: %w", err)
		}
		for _, e := range employees {
			names[e.ID] = e.Name
		}
	}

	list := make([]types.UserPunishmentItem, 0, len(rows))
	for _, row := range rows {
		item := types.UserPunishmentItem{
			ID:           row.ID,
			UserID:       row.UserID,
			Action:       row.Action,
			ActionText:   utils.TextOf(punishActionText, row.Action),
			MuteUntil:    row.MuteUntil,
			TargetID:     row.TargetID,
			Reason:       row.Reason,
			EmployeeID:   row.EmployeeID,
			EmployeeName: names[row.EmployeeID],
			ReportID:     row.ReportID,
			CreatedAt:    utils.FormatDateTime(row.CreatedAt),
		}
		if row.Action == model.PunishActionMute {
			item.MuteText = punishMuteText(row.MuteUntil)
		}
		list = append(list, item)
	}
	return &types.UserPunishmentListResponse{Total: total, List: list}, nil
}

func punishMuteText(until int64) string {
	if until == model.PermanentMuteUntil {
		return "永久"
	}
	if until <= 0 {
		return ""
	}
	return "至 " + time.Unix(until, 0).Format("2006-01-02 15:04")
}

func (s *UserService) GovernanceSummary(ctx context.Context, userID uint) (*types.UserGovernanceSummary, error) {
	user, err := s.userRepo.GetByIDUnscoped(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	count, err := s.punishRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("查询处罚记录失败: %w", err)
	}
	item := toUserItem(user, s.avatarURL(ctx, user.AvatarMediaID))
	return &types.UserGovernanceSummary{
		UserID:          item.ID,
		Nickname:        item.Nickname,
		Mobile:          item.Mobile,
		AvatarURL:       item.AvatarURL,
		Status:          item.Status,
		StatusText:      item.StatusText,
		MuteText:        item.MuteText,
		Deactivated:     item.Deactivated,
		PunishmentCount: count,
	}, nil
}
