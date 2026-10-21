package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/ratelimit"
	"app/pkg/utils"
	"app/pkg/websocket"

	"github.com/redis/go-redis/v9"
)

func (s *RealtimeService) onSend(ctx context.Context, c *websocket.Conn, msg *types.ClientMsg) {
	userID := connUserID(c)
	if userID == 0 {
		s.replyError(c, "未鉴权", msg.ClientMsgID)
		return
	}

	if err := ratelimit.AllowSliding(
		ctx,
		s.cache.Client,
		fmt.Sprintf(types.CacheUserSendRateKey, userID),
		types.UserSendRateLimit,
		types.UserSendRateWindow,
	); err != nil {
		s.replyError(c, "发送过于频繁，请稍后再试", msg.ClientMsgID)
		return
	}

	lat, lng := s.resolveConnMsgLocation(ctx, c, userID, msg.Lat, msg.Lng)
	id, err := s.messageSvc.Send(ctx, &OutgoingMessageInput{
		UserID:      userID,
		Scope:       msg.Scope,
		ContentType: msg.ContentType,
		Content:     msg.Content,
		RoomID:      msg.RoomID,
		PostID:      msg.PostID,
		ToUserID:    msg.ToUserID,
		ReplyToID:   msg.ReplyToID,
		Lat:         lat,
		Lng:         lng,
	}, msg.ClientMsgID)
	if err != nil {
		s.replyError(c, clientErrorText(ctx, s.log, err, "发送失败"), msg.ClientMsgID)
		return
	}
	s.replyAck(c, msg.ClientMsgID, id)
}

func (s *ChatHub) BroadcastChat(ctx context.Context, senderUID uint, clientMsgID string, m *model.Message, item *types.MessageItem) {
	if m == nil {
		return
	}
	scope := modelScopeString(m.Scope)
	out, _ := json.Marshal(types.ServerMsg{
		Type:        types.MsgTypeMessage,
		Scope:       scope,
		Data:        item,
		ClientMsgID: clientMsgID,
	})
	switch scope {
	case types.ScopeRoom:
		s.deliverRoom(ctx, m.RoomID, out, nil)
	case types.ScopePost:
		s.deliverPost(ctx, m.PostID, out, nil)
	case types.ScopePrivate:
		s.deliverUser(ctx, m.ToUserID, out)
		if senderUID != m.ToUserID {
			s.deliverUser(ctx, senderUID, out)
		}
		if item != nil && m.ToUserID != senderUID {
			go s.deviceSvc.NotifyPrivateOffline(context.WithoutCancel(ctx), m.ToUserID, item)
		}
	case types.ScopeNearby:
		s.deliverNearby(ctx, m.Latitude, m.Longitude, senderUID, out)
		s.deliverUser(ctx, senderUID, out)
	}
}

func (s *ChatHub) BroadcastRecall(ctx context.Context, m *model.Message, item types.MessageItem) {
	if m == nil {
		return
	}
	scope := modelScopeString(m.Scope)
	out, _ := json.Marshal(types.ServerMsg{
		Type:  types.MsgTypeRecall,
		Scope: scope,
		Data:  &item,
	})
	switch scope {
	case types.ScopeRoom:
		s.deliverRoom(ctx, m.RoomID, out, nil)
	case types.ScopePost:
		s.deliverPost(ctx, m.PostID, out, nil)
	case types.ScopePrivate:
		s.deliverUser(ctx, m.ToUserID, out)
		if m.FromUserID != m.ToUserID {
			s.deliverUser(ctx, m.FromUserID, out)
		}
	case types.ScopeNearby:
		s.deliverNearby(ctx, m.Latitude, m.Longitude, m.FromUserID, out)
		s.deliverUser(ctx, m.FromUserID, out)
	}
}

func (s *ChatHub) replyError(c *websocket.Conn, text string, clientMsgID string) {
	replyJSON(c, types.ServerMsg{Type: types.MsgTypeError, Content: text, ClientMsgID: clientMsgID})
}

func (s *ChatHub) replySubscribeError(c *websocket.Conn, scope string, roomID, postID uint, text string) {
	replyJSON(c, types.ServerMsg{
		Type:    types.MsgTypeError,
		Scope:   scope,
		Content: text,
		RoomID:  roomID,
		PostID:  postID,
	})
}

func (s *ChatHub) replyAck(c *websocket.Conn, clientMsgID string, id uint) {
	if clientMsgID == "" {
		return
	}
	replyJSON(c, types.ServerMsg{Type: types.MsgTypeAck, ClientMsgID: clientMsgID, ID: id})
}

func (s *ChatHub) DeliverToUser(ctx context.Context, userID uint, payload []byte) {
	s.deliverUser(ctx, userID, payload)
}

func (s *ChatHub) deliverUser(ctx context.Context, userID uint, payload []byte) {
	if userID == 0 || len(payload) == 0 {
		return
	}
	s.hub.SendToID(userID, payload)
	s.publish(ctx, fmt.Sprintf(types.CacheWSUserChannel, userID), payload, "user", strconv.FormatUint(uint64(userID), 10))
}

func (s *ChatHub) deliverRoom(ctx context.Context, roomID uint, payload []byte, exclude *websocket.Conn) {
	if roomID == 0 || len(payload) == 0 {
		return
	}
	topic := "room:" + strconv.FormatUint(uint64(roomID), 10)
	s.hub.SendToTopic(topic, payload, exclude)
	s.publish(ctx, fmt.Sprintf(types.CacheWSRoomChannel, roomID), payload, "room", strconv.FormatUint(uint64(roomID), 10))
}

func (s *ChatHub) deliverPost(ctx context.Context, postID uint, payload []byte, exclude *websocket.Conn) {
	if postID == 0 || len(payload) == 0 {
		return
	}
	topic := "post:" + strconv.FormatUint(uint64(postID), 10)
	s.hub.SendToTopic(topic, payload, exclude)
	s.publish(ctx, fmt.Sprintf(types.CacheWSPostChannel, postID), payload, "post", strconv.FormatUint(uint64(postID), 10))
}

type nearbyBrokerMsg struct {
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	ExcludeUID uint    `json:"ex"`
	Payload    []byte  `json:"p"`
}

func (s *ChatHub) nearbyLocalTargets(ctx context.Context, lat, lng float64, excludeUID uint) []uint {
	if s.hub.ConnCount() == 0 {
		return nil
	}
	res, err := s.cache.GeoSearch(ctx, types.CacheWSOnlineGeoIndex, &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     nearbyRadiusM,
		RadiusUnit: "m",
	}).Result()
	if err != nil {
		s.log.Warn(ctx, "附近消息 GEOSEARCH 失败，回退本地扫描", logger.Error(err))
		return s.nearbyLocalTargetsScan(excludeUID, lat, lng)
	}
	targets := make([]uint, 0, len(res))
	for _, name := range res {
		id, err := strconv.ParseUint(name, 10, 64)
		if err != nil {
			continue
		}
		userID := uint(id)
		if userID == excludeUID {
			continue
		}
		if s.userInNearbyRadius(userID, lat, lng) {
			targets = append(targets, userID)
		}
	}
	return targets
}

func (s *ChatHub) nearbyLocalTargetsScan(excludeUID uint, lat, lng float64) []uint {
	snap := s.hub.SnapshotByUser()
	targets := make([]uint, 0, 32)
	for userID, conns := range snap {
		if userID == excludeUID || len(conns) == 0 {
			continue
		}
		if connSetInNearbyRadius(conns, lat, lng) {
			targets = append(targets, userID)
		}
	}
	return targets
}

func (s *ChatHub) userInNearbyRadius(userID uint, lat, lng float64) bool {
	return connSetInNearbyRadius(s.hub.ConnsByID(userID), lat, lng)
}

func connSetInNearbyRadius(conns []*websocket.Conn, lat, lng float64) bool {
	for _, c := range conns {
		clat, clng := connLocation(c)
		if clat == 0 && clng == 0 {
			continue
		}
		if utils.Haversine(lat, lng, clat, clng) <= nearbyRadiusM {
			return true
		}
	}
	return false
}

func (s *ChatHub) deliverNearby(ctx context.Context, lat, lng float64, excludeUID uint, payload []byte) {
	if lat == 0 && lng == 0 {
		s.log.Warn(ctx, "附近消息缺少坐标，跳过广播")
		return
	}

	targets := s.nearbyLocalTargets(ctx, lat, lng, excludeUID)
	if len(targets) > 0 {
		s.hub.SendToIDs(targets, payload)
	}

	batch, err := json.Marshal(nearbyBrokerMsg{
		Lat:        lat,
		Lng:        lng,
		ExcludeUID: excludeUID,
		Payload:    payload,
	})
	if err != nil {
		s.log.Warn(ctx, "WebSocket nearby 序列化失败", logger.Error(err))
		return
	}
	s.publish(ctx, types.CacheWSNearbyChannel, batch, "nearby", "")
}

func (s *ChatHub) publish(ctx context.Context, channel string, payload []byte, kind, id string) {
	if err := s.broker.Publish(ctx, channel, payload); err != nil {
		s.log.Warn(ctx, "WebSocket broker publish 失败",
			logger.String(kind+"_id", id),
			logger.Error(err))
	}
}

func replyJSON(c *websocket.Conn, msg types.ServerMsg) {
	b, _ := json.Marshal(msg)
	_ = c.Send(b)
}

func (s *ChatHub) publishCtrl(ctx context.Context, msg types.WSCtrlMsg) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.publish(ctx, types.CacheWSCtrlChannel, payload, "ctrl", msg.Op)
}
