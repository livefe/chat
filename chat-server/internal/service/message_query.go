package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

func (s *MessageService) ListRoom(ctx context.Context, userID uint, roomID uint, lat, lng float64, beforeID uint, limit int) ([]types.MessageItem, error) {
	if roomID == 0 {
		return nil, errs.Invalid("房间ID不能为空")
	}
	lat, lng = s.roomSvc.ResolveUserLocation(ctx, userID, lat, lng)
	if err := s.roomSvc.CheckAccess(ctx, roomID, lat, lng); err != nil {
		return nil, err
	}
	list, err := s.msgRepo.ListRoom(ctx, roomID, beforeID, limit)
	if err != nil {
		s.log.Error(ctx, "查询聊天区消息失败", logger.Error(err), logger.Uint("roomId", roomID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.enrichMessages(ctx, list)
}

func (s *MessageService) ListPost(ctx context.Context, userID, postID uint, beforeID uint, limit int) ([]types.MessageItem, error) {
	if err := s.postSvc.EnsurePostMessageAccess(ctx, userID, postID); err != nil {
		return nil, err
	}
	list, err := s.msgRepo.ListPost(ctx, postID, beforeID, limit)
	if err != nil {
		s.log.Error(ctx, "查询动态消息失败", logger.Error(err), logger.Uint("postId", postID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.enrichMessages(ctx, list)
}

func (s *MessageService) ListNearby(ctx context.Context, lat, lng float64, radiusMeters float64, beforeID uint, limit int) ([]types.MessageItem, error) {
	if lat == 0 && lng == 0 {
		return nil, errs.Invalid("缺少定位信息")
	}

	const metersPerDeg = 111320.0
	dLat := radiusMeters / metersPerDeg
	cosLat := math.Cos(lat * math.Pi / 180)
	dLng := radiusMeters / metersPerDeg
	if cosLat > 1e-6 {
		dLng = radiusMeters / (metersPerDeg * cosLat)
	}
	minLat, maxLat := lat-dLat, lat+dLat
	minLng, maxLng := lng-dLng, lng+dLng
	if minLat < -90 {
		minLat = -90
	}
	if maxLat > 90 {
		maxLat = 90
	}
	if minLng < -180 {
		minLng = -180
	}
	if maxLng > 180 {
		maxLng = 180
	}

	const maxRounds = 6
	fetchSize := limit * 5
	if fetchSize < 50 {
		fetchSize = 50
	}
	filtered := make([]*model.Message, 0, limit)
	cursor := beforeID
	for round := 0; round < maxRounds && len(filtered) < limit; round++ {
		list, err := s.msgRepo.ListNearbyByBox(ctx, minLat, maxLat, minLng, maxLng, cursor, fetchSize)
		if err != nil {
			s.log.Error(ctx, "查询附近消息失败", logger.Error(err))
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		if len(list) == 0 {
			break
		}
		for _, m := range list {
			cursor = m.ID
			if m.Latitude == 0 && m.Longitude == 0 {
				continue
			}
			if utils.InCircle(m.Latitude, m.Longitude, lat, lng, radiusMeters) {
				filtered = append(filtered, m)
				if len(filtered) >= limit {
					break
				}
			}
		}
		if len(list) < fetchSize {
			break
		}
	}
	items, err := s.enrichMessages(ctx, filtered)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *MessageService) ListPrivate(ctx context.Context, userA, userB uint, beforeID uint, limit int) ([]types.MessageItem, uint, error) {
	if userA == 0 || userB == 0 {
		return nil, 0, errs.Invalid("用户ID不能为空")
	}
	clearedBefore := s.conversation.PrivateClearedBeforeID(ctx, userA, userB)
	list, err := s.msgRepo.ListPrivate(ctx, userA, userB, clearedBefore, beforeID, limit)
	if err != nil {
		s.log.Error(ctx, "查询私聊消息失败", logger.Error(err))
		return nil, 0, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.enrichMessages(ctx, list)
	if err != nil {
		return nil, 0, err
	}
	return items, clearedBefore, nil
}

func (s *MessageService) ReportTarget(ctx context.Context, reporterID, targetID uint) (uint, types.ReportContextSnapshot, error) {
	msg, err := s.msgRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, types.ReportContextSnapshot{}, errs.NotFound("消息不存在")
		}
		return 0, types.ReportContextSnapshot{}, fmt.Errorf("查询失败: %w", err)
	}
	if msg.FromUserID == reporterID {
		return 0, types.ReportContextSnapshot{}, errs.Forbidden("不能举报自己的消息")
	}
	if !reporterCanAccessMessage(reporterID, msg) {
		return 0, types.ReportContextSnapshot{}, errs.Forbidden("无权举报该消息")
	}
	snapshot := types.ReportContextSnapshot{
		MessageContent: msg.Content,
		MessageType:    msg.ContentType,
		Scope:          msg.Scope,
		RoomID:         msg.RoomID,
		SentAt:         utils.FormatDateTime(msg.CreatedAt),
	}
	if msg.Scope == model.MessageScopePrivate {
		snapshot.PeerID = msg.FromUserID
	}
	profiles, err := s.userSvc.GetProfiles(ctx, []uint{msg.FromUserID})
	if err != nil {
		return 0, types.ReportContextSnapshot{}, fmt.Errorf("查询失败: %w", err)
	}
	if profile, ok := profiles[msg.FromUserID]; ok {
		snapshot.Nickname = profile.Nickname
		snapshot.Avatar = profile.AvatarURL
	}
	return msg.FromUserID, snapshot, nil
}

func reporterCanAccessMessage(reporterID uint, msg *model.Message) bool {
	if msg.Scope == model.MessageScopePrivate {
		return msg.FromUserID == reporterID || msg.ToUserID == reporterID
	}
	return msg.Scope == model.MessageScopeNearby ||
		msg.Scope == model.MessageScopeRoom ||
		msg.Scope == model.MessageScopePost
}
