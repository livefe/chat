package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"strings"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"
)

const messageLocationPlaceNameMax = 64
const messageVoiceMaxDurationMs = 60_000

type messageLocationContent struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	PlaceName string  `json:"placeName"`
}

type messageMediaContent struct {
	MediaID    uint   `json:"mediaId"`
	URL        string `json:"url,omitempty"`
	DurationMs int    `json:"durationMs,omitempty"`
	Name       string `json:"name,omitempty"`
	SizeBytes  int64  `json:"sizeBytes,omitempty"`
	MimeType   string `json:"mimeType,omitempty"`
}

type chatBroadcaster interface {
	BroadcastChat(ctx context.Context, senderUID uint, clientMsgID string, m *model.Message, item *types.MessageItem)
	BroadcastRecall(ctx context.Context, m *model.Message, item types.MessageItem)
}

type MessageService struct {
	log          *logger.Logger
	db           *database.DB
	msgRepo      repository.MessageRepository
	roomSvc      *RoomService
	conversation *ConversationService
	userSvc      *UserService
	postSvc      *PostService
	mediaSvc     *MediaService
	sensitive    *SensitiveWordService
	broadcast    chatBroadcaster
}

func NewMessageService(
	log *logger.Logger,
	db *database.DB,
	msgRepo repository.MessageRepository,
	roomSvc *RoomService,
	conversation *ConversationService,
	userSvc *UserService,
	postSvc *PostService,
	mediaSvc *MediaService,
	sensitive *SensitiveWordService,
	broadcast chatBroadcaster,
) *MessageService {
	return &MessageService{
		log:          log,
		db:           db,
		msgRepo:      msgRepo,
		roomSvc:      roomSvc,
		conversation: conversation,
		userSvc:      userSvc,
		postSvc:      postSvc,
		mediaSvc:     mediaSvc,
		sensitive:    sensitive,
		broadcast:    broadcast,
	}
}

func (s *MessageService) broadcastRecall(ctx context.Context, m *model.Message) {
	if m == nil {
		return
	}
	s.broadcast.BroadcastRecall(ctx, m, s.MessageItem(ctx, m))
}

func prepareOwnedMediaContent(ctx context.Context, mediaSvc *MediaService, owner MediaUploader, raw string, kind uint8) (string, error) {
	mediaID, err := parseMediaIDInput(raw)
	if err != nil {
		return "", err
	}
	if mediaSvc == nil {
		return "", errs.Unavailable("媒体服务不可用", nil)
	}
	if _, err := mediaSvc.AssertOwn(ctx, owner, mediaID, kind); err != nil {
		return "", err
	}
	stored, err := json.Marshal(messageMediaContent{MediaID: mediaID})
	if err != nil {
		return "", errs.Invalid("媒体无效")
	}
	return string(stored), nil
}

func contentMediaKind(contentType uint8) (uint8, bool) {
	switch contentType {
	case model.MessageContentTypeImage:
		return model.MediaKindImage, true
	case model.MessageContentTypeVideo:
		return model.MediaKindVideo, true
	case model.MessageContentTypeVoice:
		return model.MediaKindAudio, true
	case model.MessageContentTypeFile:
		return model.MediaKindFile, true
	default:
		return 0, false
	}
}

func (s *MessageService) prepareVoiceContent(ctx context.Context, userID uint, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 1000 {
		return "", errs.Invalid("语音无效")
	}
	var voice messageMediaContent
	if err := json.Unmarshal([]byte(raw), &voice); err != nil || voice.MediaID == 0 {
		return "", errs.Invalid("语音无效")
	}
	if voice.DurationMs < 1 || voice.DurationMs > messageVoiceMaxDurationMs {
		return "", errs.Invalid("语音时长无效")
	}
	if s.mediaSvc == nil {
		return "", errs.Unavailable("媒体服务不可用", nil)
	}
	if _, err := s.mediaSvc.AssertOwn(ctx, UserUploader(userID), voice.MediaID, model.MediaKindAudio); err != nil {
		return "", err
	}
	stored, err := json.Marshal(messageMediaContent{
		MediaID:    voice.MediaID,
		DurationMs: voice.DurationMs,
	})
	if err != nil {
		return "", errs.Invalid("语音无效")
	}
	return string(stored), nil
}

func prepareOwnedFileContent(ctx context.Context, mediaSvc *MediaService, owner MediaUploader, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	var body messageMediaContent
	if err := json.Unmarshal([]byte(raw), &body); err != nil || body.MediaID == 0 {
		return "", errs.Invalid("文件无效")
	}
	mediaID := body.MediaID
	clientName := strings.TrimSpace(body.Name)
	if mediaSvc == nil {
		return "", errs.Unavailable("媒体服务不可用", nil)
	}
	asset, err := mediaSvc.AssertOwn(ctx, owner, mediaID, model.MediaKindFile)
	if err != nil {
		return "", err
	}
	name := clientName
	if name == "" {
		name = path.Base(asset.ObjectKey)
	}
	if name == "" || name == "." || name == "/" {
		name = "file" + asset.Ext
	}
	if utf8.RuneCountInString(name) > 120 {
		name = string([]rune(name)[:120])
	}
	stored, err := json.Marshal(messageMediaContent{
		MediaID:   mediaID,
		Name:      name,
		SizeBytes: asset.SizeBytes,
		MimeType:  asset.MimeType,
	})
	if err != nil {
		return "", errs.Invalid("文件无效")
	}
	return string(stored), nil
}

func parseStoredMediaID(contentType uint8, content string) (uint, error) {
	var body messageMediaContent
	if err := json.Unmarshal([]byte(content), &body); err != nil || body.MediaID == 0 {
		return 0, errs.Invalid("媒体无效")
	}
	switch contentType {
	case model.MessageContentTypeImage, model.MessageContentTypeVideo, model.MessageContentTypeVoice, model.MessageContentTypeFile:
		return body.MediaID, nil
	default:
		return 0, errs.Invalid("媒体无效")
	}
}

func normalizeLocationContent(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 1000 {
		return "", errs.Invalid("位置无效")
	}
	var loc messageLocationContent
	if err := json.Unmarshal([]byte(raw), &loc); err != nil {
		return "", errs.Invalid("位置无效")
	}
	if math.IsNaN(loc.Latitude) || math.IsInf(loc.Latitude, 0) ||
		math.IsNaN(loc.Longitude) || math.IsInf(loc.Longitude, 0) {
		return "", errs.Invalid("位置无效")
	}
	if loc.Latitude < -90 || loc.Latitude > 90 || loc.Longitude < -180 || loc.Longitude > 180 {
		return "", errs.Invalid("位置无效")
	}
	if loc.Latitude == 0 && loc.Longitude == 0 {
		return "", errs.Invalid("位置无效")
	}
	name := strings.TrimSpace(loc.PlaceName)
	if name == "" {
		name = "位置"
	}
	if utf8.RuneCountInString(name) > messageLocationPlaceNameMax {
		name = string([]rune(name)[:messageLocationPlaceNameMax])
	}
	out, err := json.Marshal(messageLocationContent{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		PlaceName: name,
	})
	if err != nil {
		return "", errs.Invalid("位置无效")
	}
	return string(out), nil
}

func conversationPreview(contentType uint8, content string) string {
	preview := messagePreview(contentType, content)
	if utf8.RuneCountInString(preview) <= 100 {
		return preview
	}
	return string([]rune(preview)[:100])
}

func messagePreview(contentType uint8, content string) string {
	switch contentType {
	case model.MessageContentTypeImage:
		return "[图片]"
	case model.MessageContentTypeVoice:
		return "[语音]"
	case model.MessageContentTypeVideo:
		return "[视频]"
	case model.MessageContentTypeFile:
		var body messageMediaContent
		if err := json.Unmarshal([]byte(content), &body); err == nil && strings.TrimSpace(body.Name) != "" {
			return "[文件] " + strings.TrimSpace(body.Name)
		}
		return "[文件]"
	case model.MessageContentTypeLocation:
		return "[位置]"
	case model.MessageContentTypePost:
		return "[动态]"
	default:
		return strings.TrimSpace(content)
	}
}

func messageItem(m *model.Message, profile types.UserProfile, hasProfile bool) types.MessageItem {
	item := types.MessageItem{
		ID:          m.ID,
		Scope:       m.Scope,
		RoomID:      m.RoomID,
		PostID:      m.PostID,
		FromUserID:  m.FromUserID,
		ToUserID:    m.ToUserID,
		ContentType: m.ContentType,
		Content:     m.Content,
		ReplyToID:   m.ReplyToID,
		CreatedAt:   m.CreatedAt.Unix(),
		Recalled:    m.Recalled != 0,
	}
	if hasProfile {
		item.FromName = utils.DisplayName(profile.Nickname)
		item.FromAvatar = profile.AvatarURL
	} else {
		item.FromName = deactivatedDisplayName()
	}
	return item
}

func replyQuotePreview(contentType uint8, content string) string {
	if contentType == model.MessageContentTypeImage ||
		contentType == model.MessageContentTypeVoice ||
		contentType == model.MessageContentTypeVideo ||
		contentType == model.MessageContentTypeLocation {
		return strings.TrimSpace(content)
	}
	if contentType == model.MessageContentTypePost {
		return messagePostQuotePreview(content)
	}
	preview := messagePreview(contentType, content)
	runes := []rune(preview)
	if len(runes) > 80 {
		return string(runes[:80]) + "…"
	}
	return preview
}

func fillReplyQuote(item *types.MessageItem, target *model.Message, profile types.UserProfile, hasProfile bool) {
	if target == nil {
		item.ReplyToContent = "原消息已删除"
		return
	}
	if hasProfile {
		item.ReplyToFromName = utils.DisplayName(profile.Nickname)
	} else {
		item.ReplyToFromName = deactivatedDisplayName()
	}
	if target.Recalled != 0 {
		item.ReplyToContent = "原消息已撤回"
		return
	}
	item.ReplyToContentType = target.ContentType
	item.ReplyToContent = replyQuotePreview(target.ContentType, target.Content)
}

func sameReplyThread(msg, target *model.Message) bool {
	if msg.Scope != target.Scope {
		return false
	}
	switch msg.Scope {
	case model.MessageScopeRoom:
		return msg.RoomID != 0 && msg.RoomID == target.RoomID
	case model.MessageScopePost:
		return msg.PostID != 0 && msg.PostID == target.PostID
	case model.MessageScopeNearby:
		return true
	case model.MessageScopePrivate:
		return (msg.FromUserID == target.FromUserID && msg.ToUserID == target.ToUserID) ||
			(msg.FromUserID == target.ToUserID && msg.ToUserID == target.FromUserID)
	default:
		return false
	}
}

func (s *MessageService) MessageItem(ctx context.Context, m *model.Message, knownReply ...*model.Message) types.MessageItem {
	items, err := s.enrichMessages(ctx, []*model.Message{m}, knownReply...)
	if err != nil || len(items) == 0 {
		if err != nil {
			s.log.Warn(ctx, "组装消息展示失败", logger.Error(err))
		}
		return messageItem(m, types.UserProfile{}, false)
	}
	return items[0]
}

func (s *MessageService) enrichMessages(ctx context.Context, list []*model.Message, knownReplies ...*model.Message) ([]types.MessageItem, error) {
	if len(list) == 0 {
		return []types.MessageItem{}, nil
	}

	replyMap := make(map[uint]*model.Message, len(knownReplies))
	for _, t := range knownReplies {
		if t == nil || t.ID == 0 {
			continue
		}
		replyMap[t.ID] = t
	}

	idSet := make(map[uint]struct{}, len(list)*2)
	replyIDs := make([]uint, 0)
	replySeen := make(map[uint]struct{})
	mediaIDs := make([]uint, 0)
	mediaSeen := make(map[uint]struct{})
	collectMedia := func(contentType uint8, content string) {
		id, err := parseStoredMediaID(contentType, content)
		if err != nil || id == 0 {
			return
		}
		if _, ok := mediaSeen[id]; ok {
			return
		}
		mediaSeen[id] = struct{}{}
		mediaIDs = append(mediaIDs, id)
	}

	for _, m := range list {
		idSet[m.FromUserID] = struct{}{}
		collectMedia(m.ContentType, m.Content)
		if m.ReplyToID == 0 {
			continue
		}
		if t := replyMap[m.ReplyToID]; t != nil {
			idSet[t.FromUserID] = struct{}{}
			collectMedia(t.ContentType, t.Content)
			continue
		}
		if _, ok := replySeen[m.ReplyToID]; ok {
			continue
		}
		replySeen[m.ReplyToID] = struct{}{}
		replyIDs = append(replyIDs, m.ReplyToID)
	}

	if len(replyIDs) > 0 {
		if targets, err := s.msgRepo.GetByIDs(ctx, replyIDs); err == nil {
			for _, t := range targets {
				replyMap[t.ID] = t
				idSet[t.FromUserID] = struct{}{}
				collectMedia(t.ContentType, t.Content)
			}
		}
	}

	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	profiles, err := s.userSvc.GetProfiles(ctx, ids)
	if err != nil {
		s.log.Error(ctx, "查询消息发送人失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	urlMap := map[uint]string{}
	if s.mediaSvc != nil {
		urlMap = s.mediaSvc.ResolveURLs(ctx, mediaIDs)
	}

	out := make([]types.MessageItem, 0, len(list))
	for i := len(list) - 1; i >= 0; i-- {
		m := list[i]
		p, ok := profiles[m.FromUserID]
		item := messageItem(m, p, ok)
		item.Content = hydrateMessageContent(m.ContentType, m.Content, urlMap)
		if m.ReplyToID > 0 {
			target := replyMap[m.ReplyToID]
			if target == nil {
				fillReplyQuote(&item, nil, types.UserProfile{}, false)
			} else {
				rp, rok := profiles[target.FromUserID]
				fillReplyQuote(&item, target, rp, rok)
				item.ReplyToContent = replyQuotePreview(target.ContentType, hydrateMessageContent(target.ContentType, target.Content, urlMap))
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func hydrateMessageContent(contentType uint8, content string, urls map[uint]string) string {
	switch contentType {
	case model.MessageContentTypeImage, model.MessageContentTypeVideo, model.MessageContentTypeVoice, model.MessageContentTypeFile:
	default:
		return content
	}
	var body messageMediaContent
	if err := json.Unmarshal([]byte(content), &body); err != nil || body.MediaID == 0 {
		return content
	}
	body.URL = urls[body.MediaID]
	out, err := json.Marshal(body)
	if err != nil {
		return content
	}
	return string(out)
}
