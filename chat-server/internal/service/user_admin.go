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
	"app/pkg/auth"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var deactivationSourceText = map[uint8]string{
	model.DeactivationSourceUser:  "用户自助",
	model.DeactivationSourceAdmin: "管理员",
}

var loginStatusText = map[uint8]string{
	0: "失败",
	1: "成功",
}

func (s *UserService) AdminList(ctx context.Context, req *types.UserListRequest) (*types.UserListResponse, error) {
	users, total, err := s.userRepo.List(ctx, repository.UserListFilter{
		ID: req.ID, Mobile: req.Mobile, Nickname: req.Nickname, Status: req.Status, Muted: req.Muted,
		StartTime: req.StartTime, EndTime: req.EndTime,
		Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询用户列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	ids := make([]uint, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	lastLogin, err := s.loginRepo.LastSuccessAt(ctx, ids)
	if err != nil {
		s.log.Error(ctx, "查询最近登录时间失败", logger.Error(err))
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}

	list := make([]types.UserItem, 0, len(users))
	urlByMedia := s.avatarURLs(ctx, users)
	for _, u := range users {
		item := toUserItem(u, urlByMedia[u.AvatarMediaID])
		if t, ok := lastLogin[u.ID]; ok {
			item.LastLoginAt = utils.FormatDateTime(t)
		}
		list = append(list, item)
	}
	return &types.UserListResponse{Total: total, List: list}, nil
}

func (s *UserService) AdminCreate(ctx context.Context, req *types.CreateUserRequest) (uint, error) {
	existing, err := s.userRepo.GetByMobile(ctx, req.Mobile)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if existing != nil {
		return 0, errs.Conflict("手机号已存在")
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		return 0, errs.Invalid("请填写昵称")
	}
	if err := assertClean(s.sensitive, nickname); err != nil {
		return 0, err
	}
	user := &model.User{Mobile: req.Mobile, Nickname: nickname, Status: req.Status}
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.log.Error(ctx, "创建用户失败", logger.Error(err))
		return 0, fmt.Errorf("创建用户失败: %w", err)
	}
	return user.ID, nil
}

func (s *UserService) AdminUpdate(ctx context.Context, req *types.UpdateUserRequest) error {
	user, err := s.adminGetActive(ctx, req.ID)
	if err != nil {
		return err
	}

	oldMobile := user.Mobile
	if req.Mobile != "" && req.Mobile != user.Mobile {
		other, err := s.userRepo.GetByMobile(ctx, req.Mobile)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		if other != nil && other.ID != user.ID {
			return errs.Conflict("手机号已被其他用户使用")
		}
		user.Mobile = req.Mobile
	}

	nickname := strings.TrimSpace(req.Nickname)
	if nickname == "" {
		return errs.Invalid("请填写昵称")
	}
	if err := assertClean(s.sensitive, nickname); err != nil {
		return err
	}
	user.Nickname = nickname
	if err := s.userRepo.UpdateMobile(ctx, user, oldMobile); err != nil {
		s.log.Error(ctx, "更新用户失败", logger.Error(err))
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}

func (s *UserService) AdminGet(ctx context.Context, id uint) (*types.AdminUserDetailResponse, error) {
	user, err := s.userRepo.GetByIDUnscoped(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	deactivationLogs, err := s.deactivationRepo.ListByUserID(ctx, id, 10)
	if err != nil {
		s.log.Error(ctx, "查询注销记录失败", logger.Error(err))
		return nil, fmt.Errorf("查询注销记录失败: %w", err)
	}
	deactItems := make([]types.UserDeactivationLogItem, 0, len(deactivationLogs))
	for _, row := range deactivationLogs {
		deactItems = append(deactItems, types.UserDeactivationLogItem{
			ID:         row.ID,
			Mobile:     utils.DesensitizeMobile(row.Mobile),
			Source:     row.Source,
			SourceText: utils.TextOf(deactivationSourceText, row.Source),
			EmployeeID: row.EmployeeID,
			Reason:     row.Reason,
			CreatedAt:  utils.FormatDateTime(row.CreatedAt),
		})
	}

	rooms, err := s.adminRoomSummaries(ctx, id)
	if err != nil {
		s.log.Error(ctx, "查询用户聊天区活动失败", logger.Uint("user_id", id), logger.Error(err))
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	item := toUserItem(user, s.avatarURL(ctx, user.AvatarMediaID))
	if last, err := s.loginRepo.LastSuccessAt(ctx, []uint{id}); err == nil {
		if t, ok := last[id]; ok {
			item.LastLoginAt = utils.FormatDateTime(t)
		}
	}
	return &types.AdminUserDetailResponse{
		UserItem:         item,
		DeactivationLogs: deactItems,
		Rooms:            rooms,
	}, nil
}

func (s *UserService) AdminLoginRecords(ctx context.Context, req *types.AdminUserSubListRequest) (*types.LoginRecordListResponse, error) {
	rows, total, err := s.loginRepo.ListByUserID(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		s.log.Error(ctx, "查询登录记录失败", logger.Uint("user_id", req.UserID), logger.Error(err))
		return nil, fmt.Errorf("查询登录记录失败: %w", err)
	}
	list := make([]types.LoginRecordItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, types.LoginRecordItem{
			ID:         row.ID,
			LoginType:  row.LoginType,
			IP:         row.IP,
			UserAgent:  row.UserAgent,
			Status:     row.Status,
			StatusText: utils.TextOf(loginStatusText, row.Status),
			Remark:     row.Remark,
			CreatedAt:  utils.FormatDateTime(row.CreatedAt),
		})
	}
	return &types.LoginRecordListResponse{Total: total, List: list}, nil
}

// AdminRevealMobile 返回完整手机号；已注销用户返回注销前的原号码。
func (s *UserService) AdminRevealMobile(ctx context.Context, id uint) (string, error) {
	user, err := s.userRepo.GetByIDUnscoped(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errs.NotFound("用户不存在")
		}
		return "", fmt.Errorf("查询用户失败: %w", err)
	}
	if !user.DeletedAt.Valid {
		return user.Mobile, nil
	}
	logs, err := s.deactivationRepo.ListByUserID(ctx, id, 1)
	if err != nil {
		return "", fmt.Errorf("查询注销记录失败: %w", err)
	}
	if len(logs) == 0 {
		return "", errs.NotFound("未找到原手机号")
	}
	return logs[0].Mobile, nil
}

func (s *UserService) adminRoomSummaries(ctx context.Context, userID uint) ([]types.AdminUserRoomSummary, error) {
	rows, err := s.msgRepo.ListUserRoomActivity(ctx, userID, 20)
	if err != nil {
		return nil, fmt.Errorf("查询聊天区活动: %w", err)
	}
	if len(rows) == 0 {
		return []types.AdminUserRoomSummary{}, nil
	}

	roomIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		roomIDs = append(roomIDs, row.RoomID)
	}
	rooms, err := s.roomRepo.GetByIDs(ctx, roomIDs)
	if err != nil {
		return nil, fmt.Errorf("查询聊天区: %w", err)
	}
	roomMap := make(map[uint]*model.Room, len(rooms))
	for _, room := range rooms {
		roomMap[room.ID] = room
	}

	out := make([]types.AdminUserRoomSummary, 0, len(rows))
	for _, row := range rows {
		name := "未知聊天区"
		if room := roomMap[row.RoomID]; room != nil {
			name = room.Name
		}
		out = append(out, types.AdminUserRoomSummary{
			RoomID:        row.RoomID,
			RoomName:      name,
			MessageCount:  row.MessageCount,
			LastMessageAt: utils.FormatDateTime(row.LastMessageAt),
		})
	}
	return out, nil
}

func (s *UserService) adminGetActive(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在或已注销")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	return user, nil
}

// PunishMeta 处罚的操作人与来源，用于写入处罚记录。
type PunishMeta struct {
	EmployeeID uint
	Reason     string
	ReportID   uint
	TargetID   uint
}

func (s *UserService) AdminMute(ctx context.Context, id uint, duration int, meta PunishMeta) error {
	user, err := s.adminGetActive(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.applyMute(ctx, user, duration, meta)
	return err
}

func (s *UserService) AdminBatchMute(ctx context.Context, req *types.BatchUserMuteRequest, meta PunishMeta) (*types.BatchUserResult, error) {
	ids := uniqueNonZeroIDs(req.IDs)
	users, err := s.userRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	result := &types.BatchUserResult{Skipped: len(ids) - len(users)}
	for _, u := range users {
		changed, err := s.applyMute(ctx, u, req.Duration, meta)
		if err != nil {
			return nil, err
		}
		if changed {
			result.Updated++
		} else {
			result.Skipped++
		}
	}
	return result, nil
}

// applyMute 解除禁言时若用户本就未被禁言则不做任何变更。
func (s *UserService) applyMute(ctx context.Context, user *model.User, duration int, meta PunishMeta) (bool, error) {
	now := time.Now()
	if duration == 0 && !isUserMuted(user, now) {
		return false, nil
	}
	user.MutedUntil = calcMutedUntil(duration, now)
	action := model.PunishActionMute
	if duration == 0 {
		action = model.PunishActionUnmute
	}
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.userRepo.Update(ctx, user); err != nil {
			return err
		}
		return s.recordPunishment(ctx, user.ID, action, user.MutedUntil, meta)
	}); err != nil {
		s.log.Error(ctx, "更新用户禁言状态失败", logger.Uint("user_id", user.ID), logger.Error(err))
		return false, fmt.Errorf("操作失败: %w", err)
	}
	return true, nil
}

func (s *UserService) AdminUpdateStatus(ctx context.Context, id uint, status uint8, meta PunishMeta) error {
	if err := validateStatusReason(status, meta.Reason); err != nil {
		return err
	}
	user, err := s.adminGetActive(ctx, id)
	if err != nil {
		return err
	}
	_, err = s.applyStatus(ctx, user, status, meta)
	return err
}

func (s *UserService) AdminBatchStatus(ctx context.Context, req *types.BatchUserStatusRequest, meta PunishMeta) (*types.BatchUserResult, error) {
	if err := validateStatusReason(req.Status, meta.Reason); err != nil {
		return nil, err
	}
	ids := uniqueNonZeroIDs(req.IDs)
	users, err := s.userRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	result := &types.BatchUserResult{Skipped: len(ids) - len(users)}
	for _, u := range users {
		changed, err := s.applyStatus(ctx, u, req.Status, meta)
		if err != nil {
			return nil, err
		}
		if changed {
			result.Updated++
		} else {
			result.Skipped++
		}
	}
	return result, nil
}

// DisableForReport 举报处置禁用账号；账号已禁用或已注销时视为成功。
func (s *UserService) DisableForReport(ctx context.Context, userID uint, meta PunishMeta) error {
	user, err := s.userRepo.GetByIDUnscoped(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if user.DeletedAt.Valid {
		return nil
	}
	_, err = s.applyStatus(ctx, user, model.UserStatusDisabled, meta)
	return err
}

func validateStatusReason(status uint8, reason string) error {
	if status == model.UserStatusDisabled && strings.TrimSpace(reason) == "" {
		return errs.Invalid("请填写禁用原因")
	}
	return nil
}

// applyStatus 状态未变化时视为成功且不重复记录。
func (s *UserService) applyStatus(ctx context.Context, user *model.User, status uint8, meta PunishMeta) (bool, error) {
	if user.Status == status {
		return false, nil
	}
	user.Status = status
	action := model.PunishActionEnable
	if status == model.UserStatusDisabled {
		action = model.PunishActionDisable
	}
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.userRepo.Update(ctx, user); err != nil {
			return err
		}
		return s.recordPunishment(ctx, user.ID, action, 0, meta)
	}); err != nil {
		s.log.Error(ctx, "更新用户状态失败", logger.Uint("user_id", user.ID), logger.Error(err))
		return false, fmt.Errorf("更新用户状态失败: %w", err)
	}
	if status != model.UserStatusEnabled {
		s.authSvc.Revoke(ctx, auth.SubjectTypeUser, user.ID)
		_ = s.deviceSvc.DeleteByUserID(ctx, user.ID)
		s.endSession(ctx, user.ID)
	}
	return true, nil
}

// AdminWarn 记录警告并通过站内系统消息告知用户。
func (s *UserService) AdminWarn(ctx context.Context, userID uint, meta PunishMeta) error {
	if _, err := s.adminGetActive(ctx, userID); err != nil {
		return err
	}
	reason := strings.TrimSpace(meta.Reason)
	if reason == "" {
		return errs.Invalid("请填写警告原因")
	}
	if s.notice == nil {
		return errs.Unavailable("站内通知未就绪", nil)
	}
	text := "【平台警告】你的行为违反了社区规范：" + reason + "。请遵守社区规范，多次违规将被禁言或封禁账号。"
	if err := s.notice.NotifyUser(ctx, userID, text); err != nil {
		s.log.Error(ctx, "发送警告通知失败", logger.Uint("user_id", userID), logger.Error(err))
		return fmt.Errorf("警告通知发送失败: %w", err)
	}
	if err := s.recordPunishment(ctx, userID, model.PunishActionWarn, 0, meta); err != nil {
		s.log.Error(ctx, "记录警告失败", logger.Uint("user_id", userID), logger.Error(err))
		return fmt.Errorf("警告已发送，但记录失败: %w", err)
	}
	return nil
}
