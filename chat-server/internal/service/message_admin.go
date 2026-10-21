package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var messageScopeText = map[uint8]string{
	model.MessageScopeNearby:  "附近",
	model.MessageScopeRoom:    "聊天区",
	model.MessageScopePrivate: "私聊",
	model.MessageScopePost:    "动态讨论",
}

func (s *MessageService) AdminList(ctx context.Context, req *types.AdminMessageListRequest) (*types.AdminMessageListResponse, error) {
	if req.Keyword != "" && req.UserID == 0 && req.PeerUserID == 0 && req.RoomID == 0 && req.PostID == 0 && req.StartTime == "" {
		return nil, errs.Invalid("关键词搜索请指定范围或时间")
	}
	list, total, err := s.msgRepo.List(ctx, repository.MessageListFilter{
		UserID: req.UserID, PeerUserID: req.PeerUserID, RoomID: req.RoomID, PostID: req.PostID, Scope: req.Scope,
		Keyword: req.Keyword, StartTime: req.StartTime, EndTime: req.EndTime,
		Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询消息列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}
	items, err := s.toAdminMessageItems(ctx, list)
	if err != nil {
		return nil, err
	}
	return &types.AdminMessageListResponse{
		Total: total,
		List:  items,
	}, nil
}

func (s *MessageService) AdminGetContext(ctx context.Context, id uint, limit int) ([]types.AdminMessageItem, error) {
	msg, err := s.msgRepo.GetByIDUnscoped(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("消息不存在")
		}
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	list, err := s.msgRepo.ListAroundID(ctx, msg, limit)
	if err != nil {
		s.log.Error(ctx, "查询消息上下文失败", logger.Error(err))
		return nil, fmt.Errorf("查询消息上下文失败: %w", err)
	}
	return s.toAdminMessageItems(ctx, list)
}

func (s *MessageService) AdminDelete(ctx context.Context, id uint) error {
	msg, err := s.msgRepo.GetByIDUnscoped(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("消息不存在")
		}
		return fmt.Errorf("查询消息失败: %w", err)
	}
	if msg.DeletedAt.Valid {
		return errs.Conflict("消息已删除")
	}
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.msgRepo.Delete(ctx, msg.ID); err != nil {
			return err
		}
		if msg.Scope == model.MessageScopePost && msg.PostID > 0 && msg.Recalled == 0 {
			return s.postSvc.DecrDiscussMessageCount(ctx, msg.PostID)
		}
		return nil
	}); err != nil {
		s.log.Error(ctx, "删除消息失败", logger.Uint("message_id", id), logger.Error(err))
		return fmt.Errorf("删除消息失败: %w", err)
	}
	return nil
}

type adminMessageRefs struct {
	userIDs  map[uint]struct{}
	roomIDs  []uint
	replyMap map[uint]*model.Message
}

func (s *MessageService) toAdminMessageItems(ctx context.Context, list []*model.Message) ([]types.AdminMessageItem, error) {
	if len(list) == 0 {
		return []types.AdminMessageItem{}, nil
	}

	refs := collectAdminMessageRefs(list)
	s.loadMissingReplyTargets(ctx, list, refs)

	userIDs := make([]uint, 0, len(refs.userIDs))
	for id := range refs.userIDs {
		userIDs = append(userIDs, id)
	}
	profiles, err := s.userSvc.GetProfiles(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}
	roomNames := s.roomNames(ctx, refs.roomIDs)
	urlMap := s.adminMessageMediaURLs(ctx, list, refs.replyMap)

	out := make([]types.AdminMessageItem, 0, len(list))
	for _, m := range list {
		out = append(out, toAdminMessageItem(m, profiles, roomNames, refs.replyMap, urlMap))
	}
	return out, nil
}

func (s *MessageService) adminMessageMediaURLs(ctx context.Context, list []*model.Message, replyMap map[uint]*model.Message) map[uint]string {
	if s.mediaSvc == nil {
		return map[uint]string{}
	}
	mediaIDs := make([]uint, 0)
	seen := make(map[uint]struct{})
	collect := func(m *model.Message) {
		if m == nil {
			return
		}
		id, err := parseStoredMediaID(m.ContentType, m.Content)
		if err != nil || id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		mediaIDs = append(mediaIDs, id)
	}
	for _, m := range list {
		collect(m)
		if m.ReplyToID > 0 {
			collect(replyMap[m.ReplyToID])
		}
	}
	return s.mediaSvc.ResolveURLs(ctx, mediaIDs)
}

func collectAdminMessageRefs(list []*model.Message) *adminMessageRefs {
	refs := &adminMessageRefs{
		userIDs:  make(map[uint]struct{}, len(list)*2),
		roomIDs:  make([]uint, 0),
		replyMap: make(map[uint]*model.Message, len(list)),
	}
	for _, m := range list {
		refs.userIDs[m.FromUserID] = struct{}{}
		if m.ToUserID > 0 {
			refs.userIDs[m.ToUserID] = struct{}{}
		}
		if m.RoomID > 0 {
			refs.roomIDs = append(refs.roomIDs, m.RoomID)
		}
		refs.replyMap[m.ID] = m
	}
	return refs
}

func (s *MessageService) loadMissingReplyTargets(ctx context.Context, list []*model.Message, refs *adminMessageRefs) {
	missing := make([]uint, 0)
	seen := make(map[uint]struct{})
	for _, m := range list {
		if m.ReplyToID == 0 {
			continue
		}
		if t := refs.replyMap[m.ReplyToID]; t != nil {
			refs.userIDs[t.FromUserID] = struct{}{}
			continue
		}
		if _, ok := seen[m.ReplyToID]; ok {
			continue
		}
		seen[m.ReplyToID] = struct{}{}
		missing = append(missing, m.ReplyToID)
	}
	if len(missing) == 0 {
		return
	}
	targets, err := s.msgRepo.GetByIDsUnscoped(ctx, missing)
	if err != nil {
		return
	}
	for _, t := range targets {
		refs.replyMap[t.ID] = t
		refs.userIDs[t.FromUserID] = struct{}{}
	}
}

func (s *MessageService) roomNames(ctx context.Context, roomIDs []uint) map[uint]string {
	rooms, _ := s.roomSvc.GetRoomsByIDs(ctx, roomIDs)
	out := make(map[uint]string, len(rooms))
	for _, room := range rooms {
		out[room.ID] = room.Name
	}
	return out
}

func toAdminMessageItem(m *model.Message, profiles map[uint]types.UserProfile, roomNames map[uint]string, replyMap map[uint]*model.Message, urlMap map[uint]string) types.AdminMessageItem {
	p, ok := profiles[m.FromUserID]
	item := types.AdminMessageItem{
		MessageItem: messageItem(m, p, ok),
		Latitude:    m.Latitude,
		Longitude:   m.Longitude,
		ScopeText:   utils.TextOf(messageScopeText, m.Scope),
		RoomName:    roomNames[m.RoomID],
	}
	item.Content = hydrateMessageContent(m.ContentType, m.Content, urlMap)
	if m.ReplyToID > 0 {
		fillAdminReplyQuote(&item, replyMap[m.ReplyToID], profiles, urlMap)
	}
	if m.DeletedAt.Valid {
		item.Deleted = true
		item.DeletedAt = utils.FormatDateTime(m.DeletedAt.Time)
	}
	if p, ok := profiles[m.ToUserID]; ok {
		item.ToName = utils.DisplayName(p.Nickname)
	} else if m.ToUserID > 0 {
		item.ToName = deactivatedDisplayName()
	}
	return item
}

func fillAdminReplyQuote(item *types.AdminMessageItem, target *model.Message, profiles map[uint]types.UserProfile, urlMap map[uint]string) {
	if target == nil {
		item.ReplyToContent = "原消息已删除"
		item.ReplyToDeleted = true
		return
	}
	if rp, ok := profiles[target.FromUserID]; ok {
		item.ReplyToFromName = utils.DisplayName(rp.Nickname)
	} else {
		item.ReplyToFromName = deactivatedDisplayName()
	}
	item.ReplyToContentType = target.ContentType
	item.ReplyToContent = replyQuotePreview(target.ContentType, hydrateMessageContent(target.ContentType, target.Content, urlMap))
	item.ReplyToRecalled = target.Recalled != 0
	item.ReplyToDeleted = target.DeletedAt.Valid
}
