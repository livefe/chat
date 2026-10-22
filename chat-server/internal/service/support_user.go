package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *SupportService) GetProfile(ctx context.Context) (*types.SupportProfileResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &types.SupportProfileResponse{
		BrandName:      cfg.BrandName,
		BrandAvatarURL: cfg.BrandAvatarURL,
		WelcomeText:    cfg.WelcomeText,
		OfflineText:    cfg.OfflineText,
		Enabled:        cfg.Enabled == 1,
	}, nil
}

func (s *SupportService) GetCurrent(ctx context.Context, userID uint) (*types.SupportSessionItem, error) {
	st, err := s.supportStateRepo.GetOpenByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		st, err = s.supportStateRepo.GetLatestByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, fmt.Errorf("查询失败: %w", err)
		}
	}
	return s.sessionItemForUser(ctx, userID, st.ConversationID)
}

func (s *SupportService) Open(ctx context.Context, userID uint, source string) (*types.SupportSessionItem, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Enabled != 1 {
		return nil, errs.Invalid("客服暂未开放")
	}
	if existing, err := s.supportStateRepo.GetOpenByUserID(ctx, userID); err == nil && existing != nil {
		return s.sessionItemForUser(ctx, userID, existing.ConversationID)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}

	now := time.Now().Unix()
	welcome := strings.TrimSpace(cfg.WelcomeText)
	var conv *model.IMConversation
	var welcomeItem *types.IMMessageItem
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		created, err := s.imSvc.CreateSupportConversation(ctx, userID, cfg.BrandName, now)
		if err != nil {
			return err
		}
		conv = created
		st := &model.IMSupportState{
			ConversationID: conv.ID,
			UserID:         userID,
			QueueAt:        now,
			Source:         source,
		}
		if err := s.supportStateRepo.Create(ctx, st); err != nil {
			return err
		}
		if welcome == "" {
			return nil
		}
		welcomeItem, err = s.writeSystemMessage(ctx, conv.ID, welcome)
		return err
	}); err != nil {
		s.log.Error(ctx, "创建客服会话失败", logger.Error(err))
		return nil, fmt.Errorf("打开会话失败: %w", err)
	}
	if welcomeItem != nil {
		s.fanoutMessage(ctx, conv.ID, model.IMPrincipalSystem, 0, welcomeItem)
	}

	s.broadcastQueueUpdate(ctx)
	return s.sessionItemForUser(ctx, userID, conv.ID)
}

func (s *SupportService) Cancel(ctx context.Context, userID, conversationID uint) error {
	_, conv, err := s.loadSupportOwnedByUser(ctx, userID, conversationID)
	if err != nil {
		return err
	}
	if conv.Status != model.IMConvStatusQueued {
		return errs.Invalid("仅排队中可取消")
	}
	now := time.Now().Unix()
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.imSvc.SetConversationStatus(ctx, conversationID, model.IMConvStatusClosed); err != nil {
			return fmt.Errorf("取消失败: %w", err)
		}
		if err := s.supportStateRepo.Update(ctx, conversationID, map[string]any{
			"close_at":     now,
			"close_reason": types.SupportCloseReasonUserCancel,
		}); err != nil {
			return fmt.Errorf("取消失败: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}
	s.broadcastQueueUpdate(ctx)
	return nil
}

func (s *SupportService) UserListMessages(ctx context.Context, userID, conversationID, beforeID uint, limit int) (*types.IMMessageListResponse, error) {
	if _, _, err := s.loadSupportOwnedByUser(ctx, userID, conversationID); err != nil {
		return nil, err
	}
	return s.listMessages(ctx, conversationID, beforeID, limit)
}

func (s *SupportService) UserSend(ctx context.Context, userID uint, req *types.SupportSendRequest) (*types.IMMessageItem, error) {
	if req == nil {
		return nil, errs.Invalid("参数错误")
	}
	_, conv, err := s.loadSupportOwnedByUser(ctx, userID, req.ConversationID)
	if err != nil {
		return nil, err
	}
	if conv.Status == model.IMConvStatusClosed {
		return nil, errs.Invalid("会话已结束")
	}
	return s.sendAs(ctx, model.IMPrincipalUser, userID, req)
}

func (s *SupportService) UserRead(ctx context.Context, userID uint, req *types.SupportReadRequest) error {
	if req == nil {
		return errs.Invalid("参数错误")
	}
	if _, _, err := s.loadSupportOwnedByUser(ctx, userID, req.ConversationID); err != nil {
		return err
	}
	if err := s.imSvc.MarkRead(ctx, req.ConversationID, model.IMPrincipalUser, userID, req.LastReadMsgID); err != nil {
		return err
	}
	_ = s.clearUserUnread(ctx, userID)
	return nil
}
