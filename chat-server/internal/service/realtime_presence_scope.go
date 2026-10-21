package service

import (
	"context"
	"encoding/json"
	"strconv"

	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/websocket"
)

func (s *ChatHub) setRoomPresence(ctx context.Context, roomID, userID uint, present bool) {
	if s.presence.SetPresence(ctx, roomPresence, roomID, userID, present) {
		s.broadcastRoomPresence(ctx, roomID)
	}
}

func (s *ChatHub) touchRoomPresence(ctx context.Context, roomID, userID uint) {
	if s.presence.TouchPresence(ctx, roomPresence, roomID, userID) {
		s.broadcastRoomPresence(ctx, roomID)
	}
}

func (s *ChatHub) RoomOnlineCount(ctx context.Context, roomID uint) (int, error) {
	return s.presence.RoomOnlineCount(ctx, roomID)
}

func (s *ChatHub) setPostPresence(ctx context.Context, postID, userID uint, present bool) {
	if s.presence.SetPresence(ctx, postPresence, postID, userID, present) {
		s.broadcastPostPresence(ctx, postID)
	}
}

func (s *ChatHub) touchPostPresence(ctx context.Context, postID, userID uint) {
	if s.presence.TouchPresence(ctx, postPresence, postID, userID) {
		s.broadcastPostPresence(ctx, postID)
	}
}

func (s *ChatHub) PostOnlineCount(ctx context.Context, postID uint) (int, error) {
	return s.presence.PostOnlineCount(ctx, postID)
}

func (s *ChatHub) refreshRoomPresenceForConn(ctx context.Context, c *websocket.Conn) {
	userID := connUserID(c)
	for rid := range connRooms(c) {
		s.touchRoomPresence(ctx, rid, userID)
	}
}

func (s *ChatHub) refreshPostPresenceForConn(ctx context.Context, c *websocket.Conn) {
	userID := connUserID(c)
	for pid := range connPosts(c) {
		s.touchPostPresence(ctx, pid, userID)
	}
}

func (s *ChatHub) ClosePost(ctx context.Context, postID uint, reason string) {
	if postID == 0 {
		return
	}
	s.closePostLocal(ctx, postID, reason)
	s.publishCtrl(ctx, types.WSCtrlMsg{
		Op:     types.WSCtrlClosePost,
		PostID: postID,
		Reason: reason,
	})
	s.presence.ClearPostPresence(ctx, postID)
}

func (s *ChatHub) CloseRoom(ctx context.Context, roomID uint, reason string) {
	if roomID == 0 {
		return
	}
	s.closeRoomLocal(ctx, roomID, reason)
	s.publishCtrl(ctx, types.WSCtrlMsg{
		Op:     types.WSCtrlCloseRoom,
		RoomID: roomID,
		Reason: reason,
	})
	s.presence.ClearRoomPresence(ctx, roomID)
}

func (s *ChatHub) closePostLocal(ctx context.Context, postID uint, reason string) {
	topic := "post:" + strconv.FormatUint(uint64(postID), 10)
	for _, c := range s.hub.ConnsByTopic(topic) {
		if !postTracked(c, postID) {
			continue
		}
		if reason != "" {
			replyJSON(c, types.ServerMsg{
				Type:    types.MsgTypeError,
				Scope:   types.ScopePost,
				PostID:  postID,
				Content: reason,
			})
		}
		s.leavePostOnConn(ctx, c, connUserID(c), postID)
	}
}

func (s *ChatHub) closeRoomLocal(ctx context.Context, roomID uint, reason string) {
	topic := "room:" + strconv.FormatUint(uint64(roomID), 10)
	for _, c := range s.hub.ConnsByTopic(topic) {
		if !roomTracked(c, roomID) {
			continue
		}
		if reason != "" {
			replyJSON(c, types.ServerMsg{
				Type:    types.MsgTypeError,
				Scope:   types.ScopeRoom,
				RoomID:  roomID,
				Content: reason,
			})
		}
		s.leaveRoomOnConn(ctx, c, connUserID(c), roomID)
	}
}

func (s *ChatHub) broadcastRoomPresence(ctx context.Context, roomID uint) {
	count, err := s.presence.RoomOnlineCount(ctx, roomID)
	if err != nil {
		s.log.Warn(ctx, "查询聊天区在线人数失败",
			logger.Uint("room_id", roomID),
			logger.Error(err))
		return
	}
	out, err := json.Marshal(types.ServerMsg{
		Type:        types.MsgTypeRoomPresence,
		RoomID:      roomID,
		OnlineCount: count,
	})
	if err != nil {
		return
	}
	s.deliverRoom(ctx, roomID, out, nil)
}

func (s *ChatHub) broadcastPostPresence(ctx context.Context, postID uint) {
	count, err := s.presence.PostOnlineCount(ctx, postID)
	if err != nil {
		s.log.Warn(ctx, "查询帖子在线人数失败",
			logger.Uint("post_id", postID),
			logger.Error(err))
		return
	}
	out, err := json.Marshal(types.ServerMsg{
		Type:        types.MsgTypePostPresence,
		Scope:       types.ScopePost,
		PostID:      postID,
		OnlineCount: count,
	})
	if err != nil {
		return
	}
	s.deliverPost(ctx, postID, out, nil)
}
