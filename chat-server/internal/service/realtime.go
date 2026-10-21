package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/logger"
	"app/pkg/websocket"
)

const (
	nearbyRadiusM = types.NearbyRadiusMeters

	AttrUserID    = "user_id"
	AttrRequestID = "request_id"
	attrRooms     = "rooms"
	attrPosts     = "posts"
	attrLat       = "lat"
	attrLng       = "lng"
)

type ChatHub struct {
	log       *logger.Logger
	cache     *cache.Cache
	broker    *websocket.RedisBroker
	hub       *websocket.Hub[uint]
	maxConns  int
	presence  *PresenceService
	deviceSvc *DeviceService
}

func NewChatHub(
	log *logger.Logger,
	cache *cache.Cache,
	cfg *websocket.Config,
	instanceID string,
	presence *PresenceService,
	deviceSvc *DeviceService,
) (*ChatHub, error) {
	broker := websocket.NewRedisBroker(cache.Client, instanceID)
	h := &ChatHub{
		log:       log,
		cache:     cache,
		broker:    broker,
		hub:       websocket.NewHub[uint](),
		maxConns:  cfg.MaxConnections,
		presence:  presence,
		deviceSvc: deviceSvc,
	}
	if err := h.broker.Subscribe(types.CacheWSChannelPatternKey, h.handleBrokerMsg); err != nil {
		_ = broker.Close(context.Background())
		return nil, fmt.Errorf("WebSocket broker 订阅失败: %w", err)
	}
	return h, nil
}

type RealtimeService struct {
	*ChatHub
	roomSvc    *RoomService
	messageSvc *MessageService
	userSvc    *UserService
	postSvc    *PostService
}

func NewRealtimeService(hub *ChatHub, roomSvc *RoomService, messageSvc *MessageService, userSvc *UserService, postSvc *PostService) *RealtimeService {
	return &RealtimeService{
		ChatHub:    hub,
		roomSvc:    roomSvc,
		messageSvc: messageSvc,
		userSvc:    userSvc,
		postSvc:    postSvc,
	}
}

func (s *ChatHub) Close() error {
	n := s.hub.CloseAll()
	if n > 0 {
		s.log.Info(context.Background(), "WebSocket 服务关闭", logger.Int("closed_conns", n))
	}
	return s.broker.Close(context.Background())
}

func (s *RealtimeService) OnOpen(c *websocket.Conn) {
	ctx := connCtx(c)
	uid := connUserID(c)
	_, err := s.userSvc.ActiveMuted(ctx, uid)
	if err != nil {
		s.log.Warn(ctx, "WebSocket 连接拒绝",
			logger.Uint("user_id", uid),
			logger.String("remote", c.RemoteAddr()),
			logger.Error(err))
		c.Close()
		return
	}
	if !s.hub.RegisterIfUnder(uid, c, s.maxConns) {
		s.log.Warn(ctx, "WebSocket 连接数已满",
			logger.Uint("user_id", uid),
			logger.Int("max_connections", s.maxConns),
			logger.Int("current", s.hub.ConnCount()))
		c.Close()
		return
	}
	if err := s.presence.MarkOnline(ctx, uid); err != nil {
		s.log.Warn(ctx, "WebSocket 上线失败", logger.Uint("user_id", uid), logger.Error(err))
	}
	s.log.Info(ctx, "WebSocket 已连接", logger.Uint("user_id", uid), logger.String("remote", c.RemoteAddr()))
}

func (s *ChatHub) DisconnectUser(userID uint) {
	if userID == 0 {
		return
	}
	s.publishCtrl(context.Background(), types.WSCtrlMsg{
		Op:     types.WSCtrlDisconnect,
		UserID: userID,
	})
	s.hub.CloseByID(userID)
	s.presence.ClearOnline(context.Background(), userID)
}

func (s *ChatHub) AtCapacity() bool {
	return s.maxConns > 0 && s.hub.ConnCount() >= s.maxConns
}

func (s *RealtimeService) OnClose(c *websocket.Conn) {
	ctx := connCtx(c)
	uid := connUserID(c)
	for rid := range connRooms(c) {
		s.leaveRoomOnConn(ctx, c, uid, rid)
	}
	for pid := range connPosts(c) {
		s.leavePostOnConn(ctx, c, uid, pid)
	}
	s.hub.Unregister(c)
	remaining := s.hub.CountByID(uid)
	if remaining == 0 {
		s.presence.ClearOnline(ctx, uid)
	}
	s.log.Info(ctx, "WebSocket 已断开",
		logger.Uint("user_id", uid),
		logger.String("remote", c.RemoteAddr()),
		logger.Int("remaining_conns", remaining))
}

func (s *RealtimeService) OnMessage(c *websocket.Conn, raw []byte) {
	var msg types.ClientMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		s.replyError(c, "消息格式错误", "")
		return
	}

	ctx := connCtx(c)
	switch msg.Type {
	case types.MsgTypePing:
		userID := connUserID(c)
		_ = s.presence.MarkOnline(ctx, userID)
		s.refreshRoomPresenceForConn(ctx, c)
		s.refreshPostPresenceForConn(ctx, c)
		s.refreshUserMuteStatus(ctx, userID)
	case types.MsgTypeSubscribe:
		s.onSubscribe(c, &msg)
	case types.MsgTypeUnsubscribe:
		s.onUnsubscribe(c, &msg)
	case types.MsgTypeLocation:
		s.onLocation(ctx, c, &msg)
	case types.MsgTypeSend:
		s.onSend(ctx, c, &msg)
	default:
		s.replyError(c, "未知的消息类型："+msg.Type, msg.ClientMsgID)
	}
}

func (s *ChatHub) handleBrokerMsg(channel string, payload []byte) {
	if channel == types.CacheWSCtrlChannel {
		s.handleCtrlMsg(payload)
		return
	}
	if channel == types.CacheWSNearbyChannel {
		var nb nearbyBrokerMsg
		if err := json.Unmarshal(payload, &nb); err != nil || len(nb.Payload) == 0 {
			return
		}
		targets := s.nearbyLocalTargets(context.Background(), nb.Lat, nb.Lng, nb.ExcludeUID)
		if len(targets) > 0 {
			s.hub.SendToIDs(targets, nb.Payload)
		}
		return
	}
	if id, ok := channelID(channel, types.CacheWSUserChannel); ok {
		s.hub.SendToID(id, payload)
		return
	}
	if id, ok := channelID(channel, types.CacheWSRoomChannel); ok {
		s.hub.SendToTopic("room:"+strconv.FormatUint(uint64(id), 10), payload, nil)
		return
	}
	if id, ok := channelID(channel, types.CacheWSPostChannel); ok {
		s.hub.SendToTopic("post:"+strconv.FormatUint(uint64(id), 10), payload, nil)
	}
}

func (s *ChatHub) handleCtrlMsg(payload []byte) {
	var msg types.WSCtrlMsg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	switch msg.Op {
	case types.WSCtrlDisconnect:
		if msg.UserID > 0 {
			s.hub.CloseByID(msg.UserID)
		}
	case types.WSCtrlClosePost:
		if msg.PostID > 0 {
			s.closePostLocal(context.Background(), msg.PostID, msg.Reason)
		}
	case types.WSCtrlCloseRoom:
		if msg.RoomID > 0 {
			s.closeRoomLocal(context.Background(), msg.RoomID, msg.Reason)
		}
	}
}

func channelID(channel, keyFmt string) (uint, bool) {
	prefix := strings.TrimSuffix(fmt.Sprintf(keyFmt, 0), "0")
	if !strings.HasPrefix(channel, prefix) {
		return 0, false
	}
	n, err := strconv.ParseUint(strings.TrimPrefix(channel, prefix), 10, 64)
	return uint(n), err == nil
}

func connCtx(c *websocket.Conn) context.Context {
	ctx := context.Background()
	if v, ok := c.Get(AttrRequestID); ok {
		if rid, ok := v.(string); ok && rid != "" {
			ctx = context.WithValue(ctx, logger.TraceIDKey, rid)
		}
	}
	return ctx
}

func connUserID(c *websocket.Conn) uint {
	v, ok := c.Get(AttrUserID)
	if !ok {
		return 0
	}
	id, _ := v.(uint)
	return id
}

func (s *RealtimeService) refreshUserMuteStatus(ctx context.Context, userID uint) {
	if userID == 0 {
		return
	}
	if _, err := s.userSvc.ActiveMuted(ctx, userID); err != nil {
		s.DisconnectUser(userID)
	}
}

func connLocation(c *websocket.Conn) (lat, lng float64) {
	if v, ok := c.Get(attrLat); ok {
		lat, _ = v.(float64)
	}
	if v, ok := c.Get(attrLng); ok {
		lng, _ = v.(float64)
	}
	return lat, lng
}

func connIDSet(c *websocket.Conn, attr string) map[uint]struct{} {
	v, ok := c.Get(attr)
	if !ok {
		return nil
	}
	m, _ := v.(map[uint]struct{})
	return m
}

func trackID(c *websocket.Conn, attr string, id uint) {
	m := connIDSet(c, attr)
	if m == nil {
		m = make(map[uint]struct{}, 4)
		c.Set(attr, m)
	}
	m[id] = struct{}{}
}

func untrackID(c *websocket.Conn, attr string, id uint) {
	m := connIDSet(c, attr)
	if m == nil {
		return
	}
	delete(m, id)
}

func idTracked(c *websocket.Conn, attr string, id uint) bool {
	_, ok := connIDSet(c, attr)[id]
	return ok
}

func connRooms(c *websocket.Conn) map[uint]struct{} { return connIDSet(c, attrRooms) }

func connPosts(c *websocket.Conn) map[uint]struct{} { return connIDSet(c, attrPosts) }

func trackRoom(c *websocket.Conn, roomID uint) { trackID(c, attrRooms, roomID) }

func untrackRoom(c *websocket.Conn, roomID uint) { untrackID(c, attrRooms, roomID) }

func roomTracked(c *websocket.Conn, roomID uint) bool { return idTracked(c, attrRooms, roomID) }

func trackPost(c *websocket.Conn, postID uint) { trackID(c, attrPosts, postID) }

func untrackPost(c *websocket.Conn, postID uint) { untrackID(c, attrPosts, postID) }

func postTracked(c *websocket.Conn, postID uint) bool { return idTracked(c, attrPosts, postID) }

func (s *ChatHub) RemoveOnlineGeo(ctx context.Context, userID uint) {
	s.presence.RemoveOnlineGeo(ctx, userID)
}
