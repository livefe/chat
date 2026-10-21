package service

import (
	"context"
	"strconv"

	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/websocket"
)

func (s *RealtimeService) onSubscribe(c *websocket.Conn, msg *types.ClientMsg) {
	switch msg.Scope {
	case types.ScopeRoom:
		if msg.RoomID == 0 {
			s.replyError(c, "订阅聊天区缺少房间ID", "")
			return
		}
		ctx := connCtx(c)
		userID := connUserID(c)
		lat, lng := s.resolveConnMsgLocation(ctx, c, userID, msg.Lat, msg.Lng)
		if err := s.roomSvc.CheckAccess(ctx, msg.RoomID, lat, lng); err != nil {
			s.replySubscribeError(c, types.ScopeRoom, msg.RoomID, 0, clientErrorText(ctx, s.log, err, "订阅失败"))
			return
		}
		wasIn := roomTracked(c, msg.RoomID)
		otherHad := s.userTracksRoom(userID, msg.RoomID, c)
		s.hub.Subscribe(c, "room:"+strconv.FormatUint(uint64(msg.RoomID), 10))
		trackRoom(c, msg.RoomID)
		if !wasIn {
			if !otherHad {
				s.setRoomPresence(ctx, msg.RoomID, userID, true)
			} else {
				s.touchRoomPresence(ctx, msg.RoomID, userID)
			}
		}
	case types.ScopePost:
		if msg.PostID == 0 {
			s.replyError(c, "订阅动态缺少动态ID", "")
			return
		}
		ctx := connCtx(c)
		userID := connUserID(c)
		if err := s.postSvc.EnsurePostMessageAccess(ctx, userID, msg.PostID); err != nil {
			s.replySubscribeError(c, types.ScopePost, 0, msg.PostID, clientErrorText(ctx, s.log, err, "订阅失败"))
			return
		}
		wasIn := postTracked(c, msg.PostID)
		otherHad := s.userTracksPost(userID, msg.PostID, c)
		s.hub.Subscribe(c, "post:"+strconv.FormatUint(uint64(msg.PostID), 10))
		trackPost(c, msg.PostID)
		if !wasIn {
			if !otherHad {
				s.setPostPresence(ctx, msg.PostID, userID, true)
			} else {
				s.touchPostPresence(ctx, msg.PostID, userID)
			}
		}
	case types.ScopeNearby, types.ScopePrivate:

	default:
		s.replyError(c, "不支持的消息范围: "+msg.Scope, "")
	}
}

func (s *RealtimeService) resolveConnMsgLocation(ctx context.Context, c *websocket.Conn, userID uint, msgLat, msgLng float64) (float64, float64) {
	lat, lng := connLocation(c)
	if lat == 0 && lng == 0 {
		lat, lng = msgLat, msgLng
	}
	return s.roomSvc.ResolveUserLocation(ctx, userID, lat, lng)
}

func (s *ChatHub) onUnsubscribe(c *websocket.Conn, msg *types.ClientMsg) {
	if msg.Scope == types.ScopeRoom && msg.RoomID > 0 {
		s.leaveRoomOnConn(connCtx(c), c, connUserID(c), msg.RoomID)
		return
	}
	if msg.Scope == types.ScopePost && msg.PostID > 0 {
		s.leavePostOnConn(connCtx(c), c, connUserID(c), msg.PostID)
	}
}

func (s *ChatHub) leaveRoomOnConn(ctx context.Context, c *websocket.Conn, userID, roomID uint) {
	wasInRoom := roomTracked(c, roomID)
	s.hub.Unsubscribe(c, "room:"+strconv.FormatUint(uint64(roomID), 10))
	untrackRoom(c, roomID)
	if wasInRoom && !s.userTracksRoom(userID, roomID, nil) {
		s.setRoomPresence(ctx, roomID, userID, false)
	}
}

func (s *ChatHub) userTracksRoom(userID, roomID uint, exclude *websocket.Conn) bool {
	if userID == 0 || roomID == 0 {
		return false
	}
	for _, c := range s.hub.ConnsByID(userID) {
		if exclude != nil && c == exclude {
			continue
		}
		if roomTracked(c, roomID) {
			return true
		}
	}
	return false
}

func (s *ChatHub) leavePostOnConn(ctx context.Context, c *websocket.Conn, userID, postID uint) {
	wasInPost := postTracked(c, postID)
	s.hub.Unsubscribe(c, "post:"+strconv.FormatUint(uint64(postID), 10))
	untrackPost(c, postID)
	if wasInPost && !s.userTracksPost(userID, postID, nil) {
		s.setPostPresence(ctx, postID, userID, false)
	}
}

func (s *ChatHub) userTracksPost(userID, postID uint, exclude *websocket.Conn) bool {
	if userID == 0 || postID == 0 {
		return false
	}
	for _, c := range s.hub.ConnsByID(userID) {
		if exclude != nil && c == exclude {
			continue
		}
		if postTracked(c, postID) {
			return true
		}
	}
	return false
}

func (s *RealtimeService) onLocation(ctx context.Context, c *websocket.Conn, msg *types.ClientMsg) {
	userID := connUserID(c)
	if msg.Lat == 0 && msg.Lng == 0 {
		s.replyError(c, "定位坐标无效", "")
		return
	}
	c.Set(attrLat, msg.Lat)
	c.Set(attrLng, msg.Lng)
	if err := s.presence.SaveOnlineLocation(ctx, userID, msg.Lat, msg.Lng); err != nil {
		s.log.Warn(ctx, "WebSocket 更新位置失败", logger.Uint("user_id", userID), logger.Error(err))
	}
	_ = s.presence.MarkOnline(ctx, userID)

	subscribed := connRooms(c)
	roomIDs := make([]uint, 0, len(subscribed))
	for rid := range subscribed {
		roomIDs = append(roomIDs, rid)
	}
	denied, err := s.roomSvc.CheckAccessBatch(ctx, roomIDs, msg.Lat, msg.Lng)
	if err != nil {
		s.log.Warn(ctx, "位置更新校验聊天区失败", logger.Uint("user_id", userID), logger.Error(err))
	} else {
		for _, rid := range roomIDs {
			if err := denied[rid]; err != nil {
				replyJSON(c, types.ServerMsg{
					Type:    types.MsgTypeError,
					Scope:   types.ScopeRoom,
					RoomID:  rid,
					Content: clientErrorText(ctx, s.log, err, "订阅失败"),
				})
				s.leaveRoomOnConn(ctx, c, userID, rid)
				continue
			}
			s.touchRoomPresence(ctx, rid, userID)
		}
	}
	s.refreshPostPresenceForConn(ctx, c)
}
