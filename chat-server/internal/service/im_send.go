package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/ratelimit"

	"gorm.io/gorm"
)

type IMDraft struct {
	ConversationID uint
	SenderKind     uint8
	SenderID       uint
	ContentType    uint8
	Raw            string
	ReplyToID      uint
	ClientMsgID    string
	Prepare        bool
	BindMedia      bool
}

func (s *IMService) ReserveOutgoing(ctx context.Context, kind uint8, senderID uint, clientMsgID string) (*types.IMMessageItem, bool, error) {
	if senderID == 0 {
		return nil, false, errs.Unauthorized("未鉴权")
	}
	rateKey := fmt.Sprintf(types.CacheUserSendRateKey, senderID)
	limit, window := types.UserSendRateLimit, types.UserSendRateWindow
	if kind == model.IMPrincipalEmployee {
		rateKey = fmt.Sprintf(types.CacheIMSendRateKey, senderID)
		limit, window = types.IMSendRateLimit, types.IMSendRateWindow
	}
	if err := ratelimit.AllowSliding(ctx, s.cache.Client, rateKey, limit, window); err != nil {
		return nil, false, errs.RateLimited("发送过于频繁，请稍后再试")
	}
	clientMsgID = strings.TrimSpace(clientMsgID)
	if utf8.RuneCountInString(clientMsgID) > maxIMClientMsgIDLen {
		return nil, false, errs.Invalid("clientMsgId 过长")
	}
	if clientMsgID == "" {
		return nil, true, nil
	}
	item, ok, err := s.clientMessage(ctx, kind, senderID, clientMsgID)
	if err != nil {
		return nil, false, err
	}
	if ok {
		return item, false, nil
	}
	return nil, true, nil
}

func (s *IMService) CommitOutgoing(ctx context.Context, d IMDraft) (*types.IMMessageItem, error) {
	content := d.Raw
	if d.Prepare {
		if d.ContentType == 0 {
			return nil, errs.Invalid("消息类型无效")
		}
		prepared, err := s.prepareOutgoingContent(ctx, d.SenderID, d.ContentType, d.Raw)
		if err != nil {
			return nil, err
		}
		content = prepared
	}
	return s.persistMessage(ctx, &model.IMMessage{
		ConversationID: d.ConversationID,
		SenderKind:     d.SenderKind,
		SenderID:       d.SenderID,
		ContentType:    d.ContentType,
		Content:        content,
		ReplyToID:      d.ReplyToID,
		ClientMsgID:    strings.TrimSpace(d.ClientMsgID),
	}, d.BindMedia)
}

func (s *IMService) Send(ctx context.Context, employeeID uint, req *types.IMSendMessageRequest) (*types.IMMessageItem, []uint, error) {
	if req == nil {
		return nil, nil, errs.Invalid("参数错误")
	}
	item, fresh, err := s.ReserveOutgoing(ctx, model.IMPrincipalEmployee, employeeID, req.ClientMsgID)
	if err != nil || !fresh {
		return item, nil, err
	}
	if _, err := s.requireActiveMember(ctx, req.ConversationID, employeeID); err != nil {
		return nil, nil, err
	}
	conv, err := s.convRepo.GetByID(ctx, req.ConversationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, errs.NotFound("会话不存在")
		}
		return nil, nil, fmt.Errorf("发送失败: %w", err)
	}
	if conv.Type != model.IMConvTypeStaffDM || conv.Status != model.IMConvStatusActive {
		return nil, nil, errs.Invalid("会话不可用")
	}
	item, err = s.CommitOutgoing(ctx, IMDraft{
		ConversationID: req.ConversationID,
		SenderKind:     model.IMPrincipalEmployee,
		SenderID:       employeeID,
		ContentType:    req.ContentType,
		Raw:            req.Content,
		ReplyToID:      req.ReplyToID,
		ClientMsgID:    req.ClientMsgID,
		Prepare:        true,
		BindMedia:      true,
	})
	if err != nil {
		return nil, nil, err
	}

	members, _ := s.memberRepo.ListActive(ctx, req.ConversationID)
	recipientIDs := make([]uint, 0, len(members))
	for _, m := range members {
		if m.PrincipalKind != model.IMPrincipalEmployee {
			continue
		}
		recipientIDs = append(recipientIDs, m.PrincipalID)
		if m.PrincipalID != employeeID {
			_ = s.IncrUnread(ctx, m.PrincipalID, req.ConversationID)
		}
	}
	s.deliverStaffMessage(ctx, recipientIDs, item)
	return item, recipientIDs, nil
}

func (s *IMService) persistMessage(ctx context.Context, msg *model.IMMessage, bindMedia bool) (*types.IMMessageItem, error) {
	if msg == nil {
		return nil, errs.Invalid("参数错误")
	}
	if msg.ReplyToID > 0 {
		reply, err := s.messageRepo.GetByID(ctx, msg.ReplyToID)
		if err != nil || reply.ConversationID != msg.ConversationID {
			return nil, errs.NotFound("回复的消息不存在")
		}
	}
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := s.messageRepo.Create(ctx, msg); err != nil {
			s.log.Error(ctx, "保存IM消息失败", logger.Error(err))
			return fmt.Errorf("发送失败: %w", err)
		}
		if bindMedia {
			if err := s.bindMedia(ctx, msg); err != nil {
				return err
			}
		}
		preview := conversationPreview(msg.ContentType, msg.Content)
		if err := s.convRepo.UpdatePreview(ctx, msg.ConversationID, msg.ID, preview, time.Now().Unix()); err != nil {
			s.log.Error(ctx, "更新IM会话预览失败", logger.Error(err))
			return fmt.Errorf("发送失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	items, err := s.toMessageItems(ctx, []*model.IMMessage{msg})
	if err != nil || len(items) == 0 {
		return nil, fmt.Errorf("发送失败: %w", err)
	}
	item := items[0]
	return &item, nil
}

func (s *IMService) clientMessage(ctx context.Context, senderKind uint8, senderID uint, clientMsgID string) (*types.IMMessageItem, bool, error) {
	existing, err := s.messageRepo.GetByClientMsgID(ctx, senderKind, senderID, clientMsgID)
	if err != nil || existing == nil {
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, fmt.Errorf("发送失败: %w", err)
		}
		return nil, false, nil
	}
	items, err := s.toMessageItems(ctx, []*model.IMMessage{existing})
	if err != nil || len(items) == 0 {
		return nil, false, fmt.Errorf("发送失败: %w", err)
	}
	item := items[0]
	return &item, true, nil
}

func (s *IMService) ListMessageItems(ctx context.Context, conversationID, beforeID uint, limit int) ([]types.IMMessageItem, error) {
	msgs, err := s.messageRepo.ListByConversation(ctx, conversationID, beforeID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.toMessageItems(ctx, msgs)
}

func (s *IMService) deliverStaffMessage(ctx context.Context, employeeIDs []uint, item *types.IMMessageItem) {
	s.deliverer.DeliverMessage(ctx, employeeIDs, item)
}

func (s *IMService) Recall(ctx context.Context, employeeID, messageID uint) (conversationID uint, recipientIDs []uint, err error) {
	msg, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, errs.NotFound("消息不存在")
		}
		return 0, nil, fmt.Errorf("撤回失败: %w", err)
	}
	isOwner := msg.SenderKind == model.IMPrincipalEmployee && msg.SenderID == employeeID
	switch checkRecall(isOwner, msg.Recalled == 1, msg.CreatedAt, time.Duration(types.IMRecallWindowSec)*time.Second) {
	case recallNotOwner:
		return 0, nil, errs.Forbidden("只能撤回自己的消息")
	case recallAlreadyDone:
		return msg.ConversationID, nil, nil
	case recallExpired:
		return 0, nil, errs.Invalid("已超过可撤回时间")
	}
	if _, err := s.requireActiveMember(ctx, msg.ConversationID, employeeID); err != nil {
		return 0, nil, err
	}
	conv, err := s.convRepo.GetByID(ctx, msg.ConversationID)
	if err != nil {
		return 0, nil, fmt.Errorf("撤回失败: %w", err)
	}
	if err := s.transaction(ctx, func(ctx context.Context) error {
		if err := s.messageRepo.MarkRecalled(ctx, messageID); err != nil {
			return err
		}
		last, err := s.messageRepo.LastVisible(ctx, msg.ConversationID)
		if err != nil {
			return err
		}
		lastID := messageID
		preview := "撤回了一条消息"
		if last != nil {
			lastID = last.ID
			preview = conversationPreview(last.ContentType, last.Content)
		}
		return s.convRepo.UpdatePreview(ctx, msg.ConversationID, lastID, preview, conv.LastActiveAt)
	}); err != nil {
		return 0, nil, fmt.Errorf("撤回失败: %w", err)
	}
	members, _ := s.memberRepo.ListActive(ctx, msg.ConversationID)
	ids := make([]uint, 0, len(members))
	for _, m := range members {
		if m.PrincipalKind == model.IMPrincipalEmployee {
			ids = append(ids, m.PrincipalID)
		}
	}
	s.deliverer.DeliverRecall(ctx, ids, msg.ConversationID, messageID)
	return msg.ConversationID, ids, nil
}

func (s *IMService) prepareOutgoingContent(ctx context.Context, ownerID uint, contentType uint8, raw string) (string, error) {
	switch contentType {
	case model.MessageContentTypeText:
		content := strings.TrimSpace(raw)
		if content == "" {
			return "", errs.Invalid("内容不能为空")
		}
		if utf8.RuneCountInString(content) > maxIMTextLen {
			content = string([]rune(content)[:maxIMTextLen])
		}
		if err := assertClean(s.sensitiveWord, content); err != nil {
			return "", err
		}
		return content, nil
	case model.MessageContentTypeImage:
		return prepareOwnedMediaContent(ctx, s.mediaSvc, ownerID, raw, model.MediaKindImage)
	case model.MessageContentTypeVideo:
		return prepareOwnedMediaContent(ctx, s.mediaSvc, ownerID, raw, model.MediaKindVideo)
	case model.MessageContentTypeFile:
		return prepareOwnedFileContent(ctx, s.mediaSvc, ownerID, raw)
	default:
		return "", errs.Invalid("不支持的消息类型")
	}
}

func (s *IMService) bindMedia(ctx context.Context, msg *model.IMMessage) error {
	if s.mediaSvc == nil || msg == nil {
		return nil
	}
	kind, ok := contentMediaKind(msg.ContentType)
	if !ok || msg.ContentType == model.MessageContentTypeVoice {
		return nil
	}
	mediaID, err := parseStoredMediaID(msg.ContentType, msg.Content)
	if err != nil || mediaID == 0 {
		return errs.Invalid("媒体无效")
	}
	_, err = s.mediaSvc.Bind(ctx, msg.SenderID, mediaID, kind, model.MediaBizIMMessage, msg.ID)
	return err
}
