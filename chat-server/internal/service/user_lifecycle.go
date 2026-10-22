package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/auth"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

type deactivationContext struct {
	Source     uint8
	EmployeeID uint
	Reason     string
	IP         string
	UserAgent  string
}

func (s *UserService) SendDeactivationCode(ctx context.Context, userID uint) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	return s.smsSvc.SendLoginCode(ctx, &types.SendCodeRequest{
		Mobile: user.Mobile,
		Source: types.SourceAppDeactivate,
	}, "", "")
}

func (s *UserService) Deactivate(ctx context.Context, userID uint, code, ip, ua string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if err := s.smsSvc.VerifyCode(ctx, user.Mobile, code, types.SourceAppDeactivate); err != nil {
		return err
	}
	if err := s.finalizeDeactivation(ctx, user, deactivationContext{
		Source:    model.DeactivationSourceUser,
		IP:        ip,
		UserAgent: ua,
	}); err != nil {
		return err
	}
	s.log.Info(ctx, "用户注销账号", logger.Uint("user_id", userID))
	return nil
}

func (s *UserService) AdminDeactivate(ctx context.Context, employeeID uint, userID uint, reason string, ip, ua string) error {
	user, err := s.adminGetActive(ctx, userID)
	if err != nil {
		return err
	}
	return s.finalizeDeactivation(ctx, user, deactivationContext{
		Source:     model.DeactivationSourceAdmin,
		EmployeeID: employeeID,
		Reason:     reason,
		IP:         ip,
		UserAgent:  ua,
	})
}

func (s *UserService) finalizeDeactivation(ctx context.Context, user *model.User, meta deactivationContext) error {
	oldMobile := user.Mobile
	user.Mobile = anonymizedMobile(user.ID)
	user.Nickname = model.DeactivatedNickname
	oldAvatar := user.AvatarMediaID
	user.AvatarMediaID = 0
	user.Status = model.UserStatusDisabled

	logEntry := &model.UserDeactivationLog{
		UserID:     user.ID,
		Mobile:     oldMobile,
		Source:     meta.Source,
		EmployeeID: meta.EmployeeID,
		Reason:     meta.Reason,
		IP:         meta.IP,
		UserAgent:  meta.UserAgent,
	}
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.userRepo.Deactivate(ctx, user, oldMobile); err != nil {
			return err
		}
		if oldAvatar > 0 && s.mediaSvc != nil {
			if err := s.mediaSvc.UnbindOne(ctx, oldAvatar, model.MediaBizAvatar, user.ID); err != nil {
				return err
			}
		}
		if err := s.deactivationRepo.Create(ctx, logEntry); err != nil {
			return err
		}
		if meta.Source == model.DeactivationSourceAdmin {
			if err := s.recordPunishment(ctx, user.ID, model.PunishActionDeactivate, 0, PunishMeta{
				EmployeeID: meta.EmployeeID,
				Reason:     meta.Reason,
			}); err != nil {
				return err
			}
		}
		if err := s.convDelRepo.DeleteByUserID(ctx, user.ID); err != nil {
			return err
		}
		if err := s.clearUserBlocksOnDeactivation(ctx, user.ID); err != nil {
			return err
		}
		return s.deviceSvc.DeleteByUserID(ctx, user.ID)
	}); err != nil {
		s.log.Error(ctx, "注销用户失败", logger.Uint("user_id", user.ID), logger.Error(err))
		return fmt.Errorf("注销失败，请稍后重试: %w", err)
	}
	s.authSvc.Revoke(ctx, auth.SubjectTypeUser, user.ID)
	s.endSession(ctx, user.ID)
	s.log.Info(ctx, "用户已注销", logger.Uint("user_id", user.ID))
	return nil
}
