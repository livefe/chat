package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *IMService) CreateSupportConversation(ctx context.Context, userID uint, title string, now int64) (*model.IMConversation, error) {
	conv := &model.IMConversation{
		Type:         model.IMConvTypeSupport,
		Status:       model.IMConvStatusQueued,
		Title:        title,
		LastActiveAt: now,
	}
	if err := s.convRepo.Create(ctx, conv); err != nil {
		return nil, err
	}
	member := &model.IMMember{
		ConversationID: conv.ID,
		PrincipalKind:  model.IMPrincipalUser,
		PrincipalID:    userID,
		Role:           model.IMMemberRoleMember,
		JoinedAt:       now,
	}
	if err := s.memberRepo.CreateBatch(ctx, []*model.IMMember{member}); err != nil {
		return nil, err
	}
	return conv, nil
}

func (s *IMService) GetConversation(ctx context.Context, id uint) (*model.IMConversation, error) {
	conv, err := s.convRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("会话不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return conv, nil
}

func (s *IMService) SetConversationStatus(ctx context.Context, id uint, status uint8) error {
	return s.convRepo.UpdateStatus(ctx, id, status)
}

func (s *IMService) SetConversationStatusIf(ctx context.Context, id uint, from, to uint8) (bool, error) {
	return s.convRepo.UpdateStatusIf(ctx, id, from, to)
}

func (s *IMService) ReactivateMember(ctx context.Context, conversationID uint, kind uint8, principalID uint, role uint8, now int64) error {
	return s.memberRepo.Reactivate(ctx, conversationID, kind, principalID, role, now)
}

func (s *IMService) LeaveMember(ctx context.Context, conversationID uint, kind uint8, principalID uint, now int64) error {
	return s.memberRepo.Leave(ctx, conversationID, kind, principalID, now)
}

func (s *IMService) MarkRead(ctx context.Context, conversationID uint, kind uint8, principalID, lastID uint) error {
	if lastID == 0 {
		conv, err := s.GetConversation(ctx, conversationID)
		if err != nil {
			return err
		}
		lastID = conv.LastMessageID
	}
	if err := s.memberRepo.UpdateRead(ctx, conversationID, kind, principalID, lastID); err != nil {
		s.log.Error(ctx, "标记IM已读失败", logger.Error(err), logger.Uint("conversationId", conversationID))
		return fmt.Errorf("标记已读失败: %w", err)
	}
	return nil
}

func (s *IMService) ListActiveMembers(ctx context.Context, conversationID uint) ([]*model.IMMember, error) {
	return s.memberRepo.ListActive(ctx, conversationID)
}

func (s *IMService) RequireMember(ctx context.Context, conversationID uint, kind uint8, principalID uint) error {
	_, err := s.memberRepo.GetActive(ctx, conversationID, kind, principalID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.Forbidden("无权访问该会话")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	return nil
}
