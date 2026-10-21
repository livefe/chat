package types

const (
	CacheUnread = cachePrefix + "unread:%d"
)

type PostMessageQuery struct {
	PostID   uint `form:"postId" binding:"required"`
	BeforeID uint `form:"beforeId"`
	Limit    int  `form:"limit" binding:"required,min=1,max=100"`
}

type RoomMessageQuery struct {
	RoomID   uint    `form:"roomId" binding:"required"`
	BeforeID uint    `form:"beforeId"`
	Limit    int     `form:"limit" binding:"required,min=1,max=100"`
	Lat      float64 `form:"lat"`
	Lng      float64 `form:"lng"`
}

type NearbyMessageQuery struct {
	Lat      float64 `form:"lat" binding:"required"`
	Lng      float64 `form:"lng" binding:"required"`
	Radius   float64 `form:"radius" binding:"required,gt=0"`
	BeforeID uint    `form:"beforeId"`
	Limit    int     `form:"limit" binding:"required,min=1,max=100"`
}

type MessageItem struct {
	ID                 uint   `json:"id"`
	Scope              uint8  `json:"scope"`
	RoomID             uint   `json:"roomId"`
	PostID             uint   `json:"postId"`
	FromUserID         uint   `json:"fromUserId"`
	ToUserID           uint   `json:"toUserId"`
	FromName           string `json:"fromName"`
	FromAvatar         string `json:"fromAvatar"`
	ContentType        uint8  `json:"contentType"`
	Content            string `json:"content"`
	ReplyToID          uint   `json:"replyToId,omitempty"`
	ReplyToFromName    string `json:"replyToFromName,omitempty"`
	ReplyToContentType uint8  `json:"replyToContentType,omitempty"`
	ReplyToContent     string `json:"replyToContent,omitempty"`
	CreatedAt          int64  `json:"createdAt"`
	Recalled           bool   `json:"recalled"`
}

type MessageListResponse struct {
	List []MessageItem `json:"list"`
}

type PrivateMessageQuery struct {
	PeerID   uint `form:"peerId" binding:"required"`
	BeforeID uint `form:"beforeId"`
	Limit    int  `form:"limit" binding:"required,min=1,max=100"`
}

type PrivateMessagesResponse struct {
	List            []MessageItem `json:"list"`
	ClearedBeforeID uint          `json:"clearedBeforeId"`
}

type AdminMessageListRequest struct {
	UserID     uint   `form:"userId"`
	PeerUserID uint   `form:"peerUserId"`
	RoomID     uint   `form:"roomId"`
	PostID     uint   `form:"postId"`
	Scope      *uint8 `form:"scope"`
	Keyword    string `form:"keyword"`
	StartTime  string `form:"startTime"`
	EndTime    string `form:"endTime"`
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminMessageItem struct {
	MessageItem
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	ScopeText       string  `json:"scopeText"`
	RoomName        string  `json:"roomName"`
	ToName          string  `json:"toName"`
	Deleted         bool    `json:"deleted"`
	DeletedAt       string  `json:"deletedAt"`
	ReplyToRecalled bool    `json:"replyToRecalled"`
	ReplyToDeleted  bool    `json:"replyToDeleted"`
}

type AdminMessageListResponse struct {
	Total int64              `json:"total"`
	List  []AdminMessageItem `json:"list"`
}

type AdminMessageContextResponse struct {
	List []AdminMessageItem `json:"list"`
}

type AdminMessageContextRequest struct {
	ID    uint `form:"id" binding:"required"`
	Limit int  `form:"limit" binding:"required,min=1,max=100"`
}
