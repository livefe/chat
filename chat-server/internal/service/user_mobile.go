package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *UserService) SendChangeMobileOldCode(ctx context.Context, userID uint) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	return s.smsSvc.SendLoginCode(ctx, &types.SendCodeRequest{
		Mobile: user.Mobile,
		Source: types.SourceAppChangeMobileOld,
	}, "", "")
}

func (s *UserService) VerifyChangeMobileOldCode(ctx context.Context, userID uint, code string) (*types.VerifyOldMobileCodeResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("用户不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if err := s.smsSvc.VerifyCode(ctx, user.Mobile, code, types.SourceAppChangeMobileOld); err != nil {
		return nil, err
	}
	ticket, err := s.smsSvc.IssueChangeMobileTicket(ctx, userID)
	if err != nil {
		s.log.Error(ctx, "签发换绑凭证失败", logger.Error(err))
		return nil, err
	}
	return &types.VerifyOldMobileCodeResponse{Ticket: ticket}, nil
}

func (s *UserService) SendChangeMobileNewCode(ctx context.Context, userID uint, newMobile, ticket string) error {
	if err := s.smsSvc.CheckChangeMobileTicket(ctx, userID, ticket); err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if newMobile == user.Mobile {
		return errs.Invalid("新手机号不能与当前手机号相同")
	}
	other, err := s.userRepo.GetByMobile(ctx, newMobile)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if other != nil {
		return errs.Conflict("该手机号已被绑定")
	}
	return s.smsSvc.SendLoginCode(ctx, &types.SendCodeRequest{
		Mobile: newMobile,
		Source: types.SourceAppChangeMobileNew,
	}, "", "")
}

func (s *UserService) ChangeMobile(ctx context.Context, userID uint, req *types.ChangeMobileRequest) error {
	if err := s.smsSvc.CheckChangeMobileTicket(ctx, userID, req.Ticket); err != nil {
		return err
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("用户不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if req.NewMobile == user.Mobile {
		return errs.Invalid("新手机号不能与当前手机号相同")
	}
	other, err := s.userRepo.GetByMobile(ctx, req.NewMobile)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if other != nil {
		return errs.Conflict("该手机号已被绑定")
	}
	if err := s.smsSvc.VerifyCode(ctx, req.NewMobile, req.NewCode, types.SourceAppChangeMobileNew); err != nil {
		return err
	}

	if err := s.userRepo.Update(ctx, userID, repository.UserUpdate{Mobile: &req.NewMobile}); err != nil {
		s.log.Error(ctx, "更换手机号失败", logger.Error(err))
		return fmt.Errorf("更换失败: %w", err)
	}
	s.smsSvc.ClearChangeMobileTicket(ctx, userID)
	s.log.Info(ctx, "用户更换手机号", logger.Uint("user_id", userID))
	return nil
}
