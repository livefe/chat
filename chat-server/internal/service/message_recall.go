package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/model"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

const recallWindow = 2 * time.Minute

type recallVerdict uint8

const (
	recallAllowed recallVerdict = iota
	recallNotOwner
	recallAlreadyDone
	recallExpired
)

func checkRecall(isOwner, recalled bool, createdAt time.Time, window time.Duration) recallVerdict {
	switch {
	case !isOwner:
		return recallNotOwner
	case recalled:
		return recallAlreadyDone
	case time.Since(createdAt) > window:
		return recallExpired
	default:
		return recallAllowed
	}
}

func (s *MessageService) Recall(ctx context.Context, userID, id uint) error {
	if id == 0 {
		return errs.Invalid("消息ID不能为空")
	}
	m, err := s.msgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("消息不存在")
		}
		s.log.Error(ctx, "查询撤回消息失败", logger.Error(err), logger.Uint("id", id))
		return fmt.Errorf("撤回失败: %w", err)
	}
	switch checkRecall(m.FromUserID == userID, m.Recalled != 0, m.CreatedAt, recallWindow) {
	case recallNotOwner:
		return errs.Forbidden("只能撤回自己的消息")
	case recallAlreadyDone:
		return errs.Conflict("消息已撤回")
	case recallExpired:
		return errs.Invalid("超过撤回时限")
	}
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.msgRepo.MarkRecalled(ctx, id); err != nil {
			return err
		}
		m.Recalled = 1
		if m.Scope == model.MessageScopePost && m.PostID > 0 {
			if err := s.postSvc.DecrDiscussMessageCount(ctx, m.PostID); err != nil {
				return err
			}
		}
		if m.Scope == model.MessageScopePrivate {
			return s.refreshPrivatePreviewAfterRecall(ctx, m)
		}
		return nil
	}); err != nil {
		var svcErr *errs.Error
		if errors.As(err, &svcErr) {
			return svcErr
		}
		s.log.Error(ctx, "撤回消息失败", logger.Error(err))
		return fmt.Errorf("撤回失败: %w", err)
	}
	s.broadcastRecall(ctx, m)
	return nil
}

func (s *MessageService) refreshPrivatePreviewAfterRecall(ctx context.Context, m *model.Message) error {
	vis, err := s.msgRepo.LastPrivateVisible(ctx, m.FromUserID, m.ToUserID)
	if err != nil {
		s.log.Warn(ctx, "撤回后查询私聊预览失败", logger.Error(err))
		return err
	}
	lastID := m.ID
	preview := "撤回了一条消息"
	if vis != nil {
		lastID = vis.ID
		preview = conversationPreview(vis.ContentType, vis.Content)
	}
	return s.conversation.updateLastPreview(ctx, m.FromUserID, m.ToUserID, lastID, preview)
}
