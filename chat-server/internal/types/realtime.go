package types

const (
	CacheWSPresenceKey       = cachePrefix + "ws:presence:user:%d"
	CacheWSOnlineGeoIndex    = cachePrefix + "geo:online"
	CacheWSUserChannel       = cachePrefix + "ws:ch:user:%d"
	CacheWSRoomChannel       = cachePrefix + "ws:ch:room:%d"
	CacheWSPostChannel       = cachePrefix + "ws:ch:post:%d"
	CacheWSNearbyChannel     = cachePrefix + "ws:ch:nearby"
	CacheWSCtrlChannel       = cachePrefix + "ws:ch:ctrl"
	CacheWSChannelPatternKey = cachePrefix + "ws:ch:*"
	CacheRoomPresenceKey     = cachePrefix + "room:presence:%d"
	CachePostPresenceKey     = cachePrefix + "post:presence:%d"
)

const (
	WSCtrlDisconnect = "disconnect"
	WSCtrlClosePost  = "close_post"
	WSCtrlCloseRoom  = "close_room"
)

type WSCtrlMsg struct {
	Op     string `json:"op"`
	UserID uint   `json:"userId,omitempty"`
	PostID uint   `json:"postId,omitempty"`
	RoomID uint   `json:"roomId,omitempty"`
	Reason string `json:"reason,omitempty"`
}

const (
	ScopeNearby  = "nearby"
	ScopePrivate = "private"
	ScopeRoom    = "room"
	ScopePost    = "post"
)

const (
	MsgTypePing        = "ping"
	MsgTypeSubscribe   = "subscribe"
	MsgTypeUnsubscribe = "unsubscribe"
	MsgTypeLocation    = "location"
	MsgTypeSend        = "send"

	MsgTypeMessage      = "message"
	MsgTypeRoomPresence = "roomPresence"
	MsgTypePostPresence = "postPresence"
	MsgTypeError        = "error"
	MsgTypeAck          = "ack"
	MsgTypeRecall       = "recall"
)

type ClientMsg struct {
	Type        string  `json:"type"`
	Scope       string  `json:"scope,omitempty"`
	RoomID      uint    `json:"roomId,omitempty"`
	PostID      uint    `json:"postId,omitempty"`
	ToUserID    uint    `json:"toUserId,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lng         float64 `json:"lng,omitempty"`
	ContentType uint8   `json:"contentType,omitempty"`
	Content     string  `json:"content,omitempty"`
	ReplyToID   uint    `json:"replyToId,omitempty"`
	ClientMsgID string  `json:"clientMsgId,omitempty"`
}

type ServerMsg struct {
	Type        string       `json:"type"`
	Scope       string       `json:"scope,omitempty"`
	Data        *MessageItem `json:"data,omitempty"`
	Content     string       `json:"content,omitempty"`
	ClientMsgID string       `json:"clientMsgId,omitempty"`
	ID          uint         `json:"id,omitempty"`
	RoomID      uint         `json:"roomId,omitempty"`
	PostID      uint         `json:"postId,omitempty"`
	OnlineCount int          `json:"onlineCount,omitempty"`
}
