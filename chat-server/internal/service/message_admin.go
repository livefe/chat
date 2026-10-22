package service

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	model.MessageScopePost:    "帖子讨论",
}

const (
	// adminKeywordDefaultWindow 仅按关键词搜索时默认的时间窗口，避免全表 LIKE。
	adminKeywordDefaultWindow = 7 * 24 * time.Hour
	// adminNearbyContextRadius 附近消息上下文的取样半径（米）。
	adminNearbyContextRadius = 3000.0
)

func (s *MessageService) AdminList(ctx context.Context, req *types.AdminMessageListRequest) (*types.AdminMessageListResponse, error) {
	startTime := req.StartTime
	scoped := req.ID > 0 || req.UserID > 0 || req.PeerUserID > 0 || req.RoomID > 0 || req.PostID > 0
	if req.Keyword != "" && !scoped && startTime == "" {
		startTime = time.Now().Add(-adminKeywordDefaultWindow).Format(time.DateTime)
	}
	list, total, err := s.msgRepo.List(ctx, repository.MessageListFilter{
		ID: req.ID, UserID: req.UserID, PeerUserID: req.PeerUserID, RoomID: req.RoomID, PostID: req.PostID,
		Scope: req.Scope, Keyword: req.Keyword, StartTime: startTime, EndTime: req.EndTime,
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

func (s *MessageService) AdminGetContext(ctx context.Context, req *types.AdminMessageContextRequest) ([]types.AdminMessageItem, error) {
	msg, err := s.msgRepo.GetByIDUnscoped(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("消息不存在")
		}
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	var nearby *repository.GeoBox
	if msg.Scope == model.MessageScopeNearby && (msg.Latitude != 0 || msg.Longitude != 0) {
		box := geoBoxAround(msg.Latitude, msg.Longitude, adminNearbyContextRadius)
		nearby = &box
	}
	list, err := s.msgRepo.ListContext(ctx, msg, nearby, req.BeforeID, req.AfterID, req.Limit)
	if err != nil {
		s.log.Error(ctx, "查询消息上下文失败", logger.Error(err))
		return nil, fmt.Errorf("查询消息上下文失败: %w", err)
	}
	return s.toAdminMessageItems(ctx, list)
}

// AdminDeleteMessages 批量删除并按需禁言发送者；已删除的消息计入跳过。
func (s *MessageService) AdminDeleteMessages(ctx context.Context, employeeID uint, req *types.AdminMessageDeleteRequest) (*types.AdminMessageDeleteResult, error) {
	ids := uniqueNonZeroIDs(req.IDs)
	msgs, err := s.msgRepo.GetByIDsUnscoped(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	result := &types.AdminMessageDeleteResult{Skipped: len(ids) - len(msgs)}
	senders := make([]uint, 0, len(msgs))
	for _, msg := range msgs {
		deleted, err := s.deleteWithRecord(ctx, msg, PunishMeta{EmployeeID: employeeID, Reason: req.Reason})
		if err != nil {
			return nil, err
		}
		if !deleted {
			result.Skipped++
			continue
		}
		result.Deleted++
		senders = append(senders, msg.FromUserID)
	}
	if req.MuteDuration != 0 && len(senders) > 0 {
		muted, err := s.userSvc.AdminBatchMute(ctx, &types.BatchUserMuteRequest{
			IDs: senders, Duration: req.MuteDuration, Reason: req.Reason,
		}, PunishMeta{EmployeeID: employeeID, Reason: req.Reason})
		if err != nil {
			return nil, err
		}
		result.Muted = muted.Updated
	}
	return result, nil
}

// DeleteForReport 举报处置删除消息；消息已被删除时视为成功。
func (s *MessageService) DeleteForReport(ctx context.Context, id uint, meta PunishMeta) error {
	msg, err := s.msgRepo.GetByIDUnscoped(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("消息不存在")
		}
		return fmt.Errorf("查询消息失败: %w", err)
	}
	_, err = s.deleteWithRecord(ctx, msg, meta)
	return err
}

func (s *MessageService) deleteWithRecord(ctx context.Context, msg *model.Message, meta PunishMeta) (bool, error) {
	if msg.DeletedAt.Valid {
		return false, nil
	}
	meta.TargetID = msg.ID
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.msgRepo.Delete(ctx, msg.ID); err != nil {
			return err
		}
		if msg.Scope == model.MessageScopePost && msg.PostID > 0 && msg.Recalled == 0 {
			if err := s.postSvc.DecrDiscussMessageCount(ctx, msg.PostID); err != nil {
				return err
			}
		}
		return s.userSvc.RecordContentPunishment(ctx, msg.FromUserID, model.PunishActionDeleteMessage, meta)
	}); err != nil {
		s.log.Error(ctx, "删除消息失败", logger.Uint("message_id", msg.ID), logger.Error(err))
		return false, fmt.Errorf("删除消息失败: %w", err)
	}
	return true, nil
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
	if ok {
		item.FromMobile = utils.DesensitizeMobile(p.Mobile)
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
		item.ToMobile = utils.DesensitizeMobile(p.Mobile)
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
