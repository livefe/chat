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
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var deactivationSourceText = map[uint8]string{
	model.DeactivationSourceUser:  "用户自助",
	model.DeactivationSourceAdmin: "管理员",
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

	list := make([]types.UserItem, 0, len(users))
	urlByMedia := s.avatarURLs(ctx, users)
	for _, u := range users {
		list = append(list, toUserItem(u, urlByMedia[u.AvatarMediaID]))
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
	user, err := s.userRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}

	if req.Mobile != user.Mobile {
		other, err := s.userRepo.GetByMobile(ctx, req.Mobile)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		if other != nil && other.ID != user.ID {
			return errs.Conflict("手机号已被其他用户使用")
		}
	}

	user.Mobile = req.Mobile
	nickname := strings.TrimSpace(req.Nickname)
	if nickname != "" {
		if err := assertClean(s.sensitive, nickname); err != nil {
			return err
		}
	}
	user.Nickname = nickname
	prevStatus := user.Status
	user.Status = req.Status
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error(ctx, "更新用户失败", logger.Error(err))
		return fmt.Errorf("更新用户失败: %w", err)
	}
	if req.Status != model.UserStatusEnabled && prevStatus == model.UserStatusEnabled {
		s.authSvc.Revoke(ctx, auth.SubjectTypeUser, req.ID)
		_ = s.deviceSvc.DeleteByUserID(ctx, req.ID)
	}
	if req.Status != model.UserStatusEnabled {
		s.endSession(ctx, req.ID)
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

	loginRecords, err := s.loginRepo.ListByUserID(ctx, id, 20)
	if err != nil {
		s.log.Error(ctx, "查询登录记录失败", logger.Error(err))
		return nil, fmt.Errorf("查询登录记录失败: %w", err)
	}
	deactivationLogs, err := s.deactivationRepo.ListByUserID(ctx, id, 10)
	if err != nil {
		s.log.Error(ctx, "查询注销记录失败", logger.Error(err))
		return nil, fmt.Errorf("查询注销记录失败: %w", err)
	}

	loginItems := make([]types.LoginRecordItem, 0, len(loginRecords))
	for _, row := range loginRecords {
		statusText := "失败"
		if row.Status == 1 {
			statusText = "成功"
		}
		loginItems = append(loginItems, types.LoginRecordItem{
			ID:         row.ID,
			LoginType:  row.LoginType,
			IP:         row.IP,
			Status:     row.Status,
			StatusText: statusText,
			Remark:     row.Remark,
			CreatedAt:  utils.FormatDateTime(row.CreatedAt),
		})
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

	return &types.AdminUserDetailResponse{
		UserItem:         toUserItem(user, s.avatarURL(ctx, user.AvatarMediaID)),
		LoginRecords:     loginItems,
		DeactivationLogs: deactItems,
		Rooms:            rooms,
	}, nil
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

func (s *UserService) AdminMute(ctx context.Context, id uint, duration int, reason string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}

	user.MutedUntil = calcMutedUntil(duration, time.Now())
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error(ctx, "更新用户禁言状态失败", logger.Uint("user_id", id), logger.Error(err))
		return fmt.Errorf("操作失败: %w", err)
	}

	action := "解除禁言"
	if duration != 0 {
		action = "设置禁言"
	}
	s.log.Info(ctx, "管理员"+action,
		logger.Uint("user_id", id),
		logger.Int("duration_hours", duration),
		logger.String("reason", reason),
	)
	return nil
}

func (s *UserService) AdminUpdateStatus(ctx context.Context, id uint, status uint8) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	user.Status = status
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.log.Error(ctx, "更新用户状态失败", logger.Error(err))
		return fmt.Errorf("更新用户状态失败: %w", err)
	}
	if status != model.UserStatusEnabled {
		s.authSvc.Revoke(ctx, auth.SubjectTypeUser, id)
		_ = s.deviceSvc.DeleteByUserID(ctx, id)
		s.endSession(ctx, id)
	}
	return nil
}
