package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
)

const maxMessageTextLen = 1000

type OutgoingMessageInput struct {
	UserID      uint
	Scope       string
	ContentType uint8
	Content     string
	RoomID      uint
	PostID      uint
	ToUserID    uint
	ReplyToID   uint
	Lat         float64
	Lng         float64
}

func (s *MessageService) PrepareOutgoing(ctx context.Context, in *OutgoingMessageInput) (*model.Message, *model.Message, error) {
	if in.ContentType == 0 {
		return nil, nil, errs.Invalid("消息类型无效")
	}
	content, err := s.prepareContent(ctx, in.UserID, in.ContentType, in.Content)
	if err != nil {
		return nil, nil, err
	}

	scope, err := scopeToModel(in.Scope)
	if err != nil {
		return nil, nil, err
	}
	if in.ContentType == model.MessageContentTypePost && scope == model.MessageScopePost {
		return nil, nil, errs.Invalid("不能转发到动态讨论")
	}

	m := &model.Message{
		Scope:       scope,
		FromUserID:  in.UserID,
		ContentType: in.ContentType,
		Content:     content,
	}
	if err := s.applyTarget(ctx, m, in); err != nil {
		return nil, nil, err
	}

	reply, err := s.ResolveReplyTarget(ctx, m, in.ReplyToID)
	if err != nil {
		return nil, nil, err
	}
	return m, reply, nil
}

func (s *MessageService) prepareContent(ctx context.Context, userID uint, contentType uint8, raw string) (string, error) {
	switch contentType {
	case model.MessageContentTypeText:
		content := strings.TrimSpace(raw)
		if content == "" {
			return "", errs.Invalid("内容不能为空")
		}
		if utf8.RuneCountInString(content) > maxMessageTextLen {
			content = string([]rune(content)[:maxMessageTextLen])
		}
		if err := assertClean(s.sensitive, content); err != nil {
			return "", err
		}
		return content, nil
	case model.MessageContentTypeImage:
		return prepareOwnedMediaContent(ctx, s.mediaSvc, userID, raw, model.MediaKindImage)
	case model.MessageContentTypeVoice:
		return s.prepareVoiceContent(ctx, userID, raw)
	case model.MessageContentTypeVideo:
		stored, err := prepareOwnedMediaContent(ctx, s.mediaSvc, userID, raw, model.MediaKindVideo)
		if err != nil {
			return "", errs.Invalid("视频无效")
		}
		return stored, nil
	case model.MessageContentTypeFile:
		return prepareOwnedFileContent(ctx, s.mediaSvc, userID, raw)
	case model.MessageContentTypeLocation:
		loc, err := normalizeLocationContent(raw)
		if err != nil {
			return "", err
		}
		if err := s.assertLocationPlaceClean(loc); err != nil {
			return "", err
		}
		return loc, nil
	case model.MessageContentTypePost:
		return s.normalizePostContent(ctx, raw)
	default:
		return "", errs.Invalid("不支持的消息类型")
	}
}

func (s *MessageService) applyTarget(ctx context.Context, m *model.Message, in *OutgoingMessageInput) error {
	switch in.Scope {
	case types.ScopeRoom:
		if in.RoomID == 0 {
			return errs.Invalid("发送聊天区消息缺少房间ID")
		}
		if err := s.roomSvc.CheckAccess(ctx, in.RoomID, in.Lat, in.Lng); err != nil {
			return err
		}
		m.RoomID = in.RoomID
	case types.ScopePost:
		if in.PostID == 0 {
			return errs.Invalid("发送动态消息缺少动态ID")
		}
		if err := s.postSvc.EnsurePostMessageAccess(ctx, in.UserID, in.PostID); err != nil {
			return err
		}
		m.PostID = in.PostID
	case types.ScopePrivate:
		if in.ToUserID == 0 {
			return errs.Invalid("私聊缺少或无效的对方用户")
		}
		if err := s.userSvc.EnsureActive(ctx, in.ToUserID); err != nil {
			return err
		}
		if err := s.userSvc.AssertCanPrivateChat(ctx, in.UserID, in.ToUserID); err != nil {
			return err
		}
		m.ToUserID = in.ToUserID
	case types.ScopeNearby:
		if in.Lat == 0 && in.Lng == 0 {
			return errs.Invalid("缺少定位信息，请刷新位置后重试")
		}
		m.Latitude, m.Longitude = in.Lat, in.Lng
	}
	return nil
}

func (s *MessageService) assertLocationPlaceClean(raw string) error {
	var loc messageLocationContent
	if err := json.Unmarshal([]byte(raw), &loc); err != nil {
		return nil
	}
	return assertClean(s.sensitive, loc.PlaceName)
}

func (s *MessageService) Send(ctx context.Context, in *OutgoingMessageInput, clientMsgID string) (uint, error) {
	if in == nil {
		return 0, errs.Invalid("参数错误")
	}
	if err := s.assertNotMuted(ctx, in.UserID); err != nil {
		return 0, err
	}
	m, reply, err := s.PrepareOutgoing(ctx, in)
	if err != nil {
		return 0, err
	}
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.PersistOutgoing(ctx, m); err != nil {
			return err
		}
		if m.Scope != model.MessageScopePrivate {
			return nil
		}
		if err := s.conversation.upsert(ctx, in.UserID, m.ToUserID, m); err != nil {
			s.log.Error(ctx, "私聊会话更新失败", logger.Error(err))
			return fmt.Errorf("发送失败: %w", err)
		}
		return nil
	}); err != nil {
		return 0, err
	}
	if m.Scope == model.MessageScopePrivate {
		if err := s.conversation.IncrUnread(ctx, m.ToUserID, in.UserID); err != nil {
			s.log.Warn(ctx, "私聊未读增加失败", logger.Error(err))
		}
	}
	item := s.MessageItem(ctx, m, reply)
	s.broadcast.BroadcastChat(ctx, in.UserID, clientMsgID, m, &item)
	return m.ID, nil
}

func (s *MessageService) assertNotMuted(ctx context.Context, userID uint) error {
	muted, err := s.userSvc.ActiveMuted(ctx, userID)
	if err != nil {
		return err
	}
	if muted {
		return errs.Forbidden("您已被禁言，暂时无法发言")
	}
	return nil
}

func (s *MessageService) PersistOutgoing(ctx context.Context, m *model.Message) error {
	if err := s.msgRepo.Create(ctx, m); err != nil {
		s.log.Error(ctx, "保存消息失败",
			logger.String("scope", modelScopeString(m.Scope)), logger.Error(err))
		return fmt.Errorf("发送失败: %w", err)
	}
	if err := s.bindMessageMedia(ctx, m); err != nil {
		s.log.Error(ctx, "绑定消息媒体失败", logger.Uint("message_id", m.ID), logger.Error(err))
		return err
	}
	if m.Scope == model.MessageScopePost && m.PostID > 0 {
		if err := s.postSvc.IncrDiscussMessageCount(ctx, m.PostID); err != nil {
			return fmt.Errorf("发送失败: %w", err)
		}
	}
	return nil
}

func (s *MessageService) bindMessageMedia(ctx context.Context, m *model.Message) error {
	if s.mediaSvc == nil || m == nil {
		return nil
	}
	kind, ok := contentMediaKind(m.ContentType)
	if !ok {
		return nil
	}
	mediaID, err := parseStoredMediaID(m.ContentType, m.Content)
	if err != nil || mediaID == 0 {
		return errs.Invalid("媒体无效")
	}
	_, err = s.mediaSvc.Bind(ctx, m.FromUserID, mediaID, kind, model.MediaBizMessage, m.ID)
	return err
}

func (s *MessageService) ResolveReplyTarget(ctx context.Context, draft *model.Message, replyToID uint) (*model.Message, error) {
	if replyToID == 0 {
		return nil, nil
	}
	target, err := s.msgRepo.GetByID(ctx, replyToID)
	if err != nil {
		return nil, errs.NotFound("回复的消息不存在")
	}
	if !sameReplyThread(draft, target) {
		return nil, errs.Invalid("只能回复同一会话中的消息")
	}
	if target.Recalled != 0 {
		return nil, errs.Invalid("不能回复已撤回的消息")
	}
	draft.ReplyToID = replyToID
	return target, nil
}

func modelScopeString(scope uint8) string {
	switch scope {
	case model.MessageScopeNearby:
		return types.ScopeNearby
	case model.MessageScopeRoom:
		return types.ScopeRoom
	case model.MessageScopePrivate:
		return types.ScopePrivate
	case model.MessageScopePost:
		return types.ScopePost
	default:
		return ""
	}
}

func scopeToModel(scope string) (uint8, error) {
	switch scope {
	case types.ScopeNearby:
		return model.MessageScopeNearby, nil
	case types.ScopeRoom:
		return model.MessageScopeRoom, nil
	case types.ScopePrivate:
		return model.MessageScopePrivate, nil
	case types.ScopePost:
		return model.MessageScopePost, nil
	default:
		return 0, errs.Invalid("不支持的消息范围: " + scope)
	}
}

func parseMediaIDInput(raw string) (uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errs.Invalid("媒体无效")
	}
	if strings.HasPrefix(raw, "{") {
		var body struct {
			MediaID uint `json:"mediaId"`
		}
		if err := json.Unmarshal([]byte(raw), &body); err != nil || body.MediaID == 0 {
			return 0, errs.Invalid("媒体无效")
		}
		return body.MediaID, nil
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errs.Invalid("媒体无效")
	}
	return uint(id), nil
}
