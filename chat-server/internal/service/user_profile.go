package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

func (s *UserService) Get(ctx context.Context, userID uint) (*types.UserInfo, error) {
	user, err := s.userRepo.GetCached(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if user.Status != model.UserStatusEnabled {
		return nil, errs.Forbidden("用户已被禁用")
	}
	info := toUserInfo(user, s.avatarURL(ctx, user.AvatarMediaID))
	return &info, nil
}

func (s *UserService) ActiveMuted(ctx context.Context, userID uint) (muted bool, err error) {
	user, err := s.userRepo.GetCached(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errs.NotFound("用户不存在")
		}
		return false, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if user.Status != model.UserStatusEnabled {
		return false, errs.Forbidden("用户已被禁用")
	}
	return isUserMuted(user, time.Now()), nil
}

func (s *UserService) PublicProfile(ctx context.Context, viewerID, targetID uint) (*types.PublicUserProfile, error) {
	if targetID == 0 {
		return nil, errs.NotFound("用户不存在")
	}
	if viewerID == targetID {
		user, err := s.userRepo.GetCached(ctx, targetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errs.NotFound("用户不存在")
			}
			return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		if user.Status != model.UserStatusEnabled {
			return nil, errs.Forbidden("用户已被禁用")
		}
		info := toUserInfo(user, s.avatarURL(ctx, user.AvatarMediaID))
		return &types.PublicUserProfile{
			ID:         info.ID,
			Nickname:   info.Nickname,
			AvatarURL:  info.AvatarURL,
			CreatedAt:  utils.FormatDateTime(user.CreatedAt),
			IsSelf:     true,
			CanMessage: true,
		}, nil
	}

	user, err := s.userRepo.GetCached(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if user.Nickname == model.DeactivatedNickname {
		return &types.PublicUserProfile{
			ID:          user.ID,
			Nickname:    deactivatedDisplayName(),
			AvatarURL:   "",
			CreatedAt:   utils.FormatDateTime(user.CreatedAt),
			Deactivated: true,
			CanMessage:  false,
		}, nil
	}
	if user.Status != model.UserStatusEnabled {
		return nil, errs.NotFound("用户不存在")
	}

	blockedByMe := false
	canMessage := true
	if viewerID > 0 {
		blockedByMe, err = s.IsBlockedBy(ctx, viewerID, targetID)
		if err != nil {
			return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		blockedEither, err := s.IsBlockedEither(ctx, viewerID, targetID)
		if err != nil {
			return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		canMessage = !blockedEither
	}

	return &types.PublicUserProfile{
		ID:          user.ID,
		Nickname:    utils.DisplayName(user.Nickname),
		AvatarURL:   s.avatarURL(ctx, user.AvatarMediaID),
		CreatedAt:   utils.FormatDateTime(user.CreatedAt),
		BlockedByMe: blockedByMe,
		CanMessage:  canMessage,
	}, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, req *types.UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	var upd repository.UserUpdate
	if req.Nickname != "" {
		if err := assertClean(s.sensitive, req.Nickname); err != nil {
			return err
		}
		upd.Nickname = &req.Nickname
	}
	if req.AvatarMediaID != nil {
		upd.AvatarMediaID = req.AvatarMediaID
	}
	persist := func() error {
		if err := s.userRepo.Update(ctx, userID, upd); err != nil {
			s.log.Error(ctx, "更新用户资料失败", logger.Error(err))
			return fmt.Errorf("更新失败: %w", err)
		}
		return nil
	}
	if req.AvatarMediaID == nil {
		return persist()
	}
	if s.mediaSvc == nil {
		return errs.Unavailable("媒体服务不可用", nil)
	}
	return s.mediaSvc.ReplaceSingle(ctx, UserUploader(userID), user.AvatarMediaID, *req.AvatarMediaID, model.MediaKindImage, model.MediaBizAvatar, userID, persist)
}

func (s *UserService) ReportTarget(ctx context.Context, reporterID, targetID uint) (uint, types.ReportContextSnapshot, error) {
	if targetID == reporterID {
		return 0, types.ReportContextSnapshot{}, errs.Forbidden("不能举报自己")
	}
	user, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, types.ReportContextSnapshot{}, errs.NotFound("用户不存在")
		}
		return 0, types.ReportContextSnapshot{}, fmt.Errorf("查询失败: %w", err)
	}
	if user.Status != model.UserStatusEnabled || user.Nickname == model.DeactivatedNickname {
		return 0, types.ReportContextSnapshot{}, errs.NotFound("用户不存在")
	}
	return user.ID, types.ReportContextSnapshot{
		Nickname: user.Nickname,
		Avatar:   s.avatarURL(ctx, user.AvatarMediaID),
	}, nil
}

func (s *UserService) GetProfiles(ctx context.Context, ids []uint) (map[uint]types.UserProfile, error) {
	if len(ids) == 0 {
		return map[uint]types.UserProfile{}, nil
	}
	users, err := s.userRepo.GetManyCached(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("调用 GetByIDsWithCache 失败: %w", err)
	}
	urlByMedia := s.avatarURLs(ctx, users)
	out := make(map[uint]types.UserProfile, len(users))
	for _, u := range users {
		if u == nil {
			continue
		}
		out[u.ID] = types.UserProfile{
			ID:        u.ID,
			Mobile:    u.Mobile,
			Nickname:  u.Nickname,
			AvatarURL: urlByMedia[u.AvatarMediaID],
			Status:    u.Status,
		}
	}
	return out, nil
}
