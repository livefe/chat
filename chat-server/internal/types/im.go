package types

const (
	CacheIMUnreadKey                = cachePrefix + "im:unread:%d"
	CacheWSEmployeeChannel          = cachePrefix + "ws:ch:employee:%d"
	CacheWSEmployeePatternKey       = cachePrefix + "ws:ch:employee:*"
	CacheWSEmployeeBroadcastChannel = cachePrefix + "ws:ch:employee:all"
	CacheWSEmployeePresenceKey      = cachePrefix + "ws:presence:employee:%d"

	IMRecallWindowSec = 120

	IMMsgTypePing         = "ping"
	IMMsgTypeSend         = "send"
	IMMsgTypeMessage      = "message"
	IMMsgTypeAck          = "ack"
	IMMsgTypeError        = "error"
	IMMsgTypeRecall       = "recall"
	IMMsgTypeSupportQueue = "support.queue"
)

type IMContactQuery struct {
	PageQuery
	Keyword      string `form:"keyword"`
	DepartmentID uint   `form:"departmentId"`
}

type IMContactItem struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	AvatarURL      string `json:"avatarUrl"`
	DepartmentID   uint   `json:"departmentId"`
	DepartmentName string `json:"departmentName"`
	Online         bool   `json:"online"`
}

type IMContactListResponse struct {
	Total int64           `json:"total"`
	List  []IMContactItem `json:"list"`
}

type IMOpenDMRequest struct {
	PeerEmployeeID uint `json:"peerEmployeeId" binding:"required"`
}

type IMConversationItem struct {
	ID            uint   `json:"id"`
	Type          uint8  `json:"type"`
	Title         string `json:"title"`
	PeerID        uint   `json:"peerId"`
	PeerName      string `json:"peerName"`
	PeerAvatarURL string `json:"peerAvatarUrl"`
	LastMessageID uint   `json:"lastMessageId"`
	LastContent   string `json:"lastContent"`
	LastActiveAt  int64  `json:"lastActiveAt"`
	UnreadCount   int64  `json:"unreadCount"`
	Pinned        bool   `json:"pinned"`
}

type IMConversationListResponse struct {
	Total int64                `json:"total"`
	List  []IMConversationItem `json:"list"`
}

type IMMessageQuery struct {
	ConversationID uint `form:"conversationId" binding:"required"`
	BeforeID       uint `form:"beforeId"`
	Limit          int  `form:"limit" binding:"required,min=1,max=100"`
}

type IMSendMessageRequest struct {
	ConversationID uint   `json:"conversationId" binding:"required"`
	ContentType    uint8  `json:"contentType"`
	Content        string `json:"content" binding:"required"`
	ReplyToID      uint   `json:"replyToId"`
	ClientMsgID    string `json:"clientMsgId"`
}

type IMReadRequest struct {
	ConversationID uint `json:"conversationId" binding:"required"`
	LastReadMsgID  uint `json:"lastReadMsgId"`
}

type IMRecallRequest struct {
	MessageID uint `json:"messageId" binding:"required"`
}

type IMMessageItem struct {
	ID                 uint   `json:"id"`
	ConversationID     uint   `json:"conversationId"`
	SenderKind         uint8  `json:"senderKind"`
	SenderID           uint   `json:"senderId"`
	SenderName         string `json:"senderName"`
	SenderAvatarURL    string `json:"senderAvatarUrl"`
	ContentType        uint8  `json:"contentType"`
	Content            string `json:"content"`
	ReplyToID          uint   `json:"replyToId,omitempty"`
	ReplyToSenderName  string `json:"replyToSenderName,omitempty"`
	ReplyToContentType uint8  `json:"replyToContentType,omitempty"`
	ReplyToContent     string `json:"replyToContent,omitempty"`
	CreatedAt          int64  `json:"createdAt"`
	Recalled           bool   `json:"recalled"`
	ClientMsgID        string `json:"clientMsgId,omitempty"`
}

type IMMessageListResponse struct {
	List []IMMessageItem `json:"list"`
}

type IMUnreadSummaryResponse struct {
	Total        int64 `json:"total"`
	SupportTotal int64 `json:"supportTotal"`
}

type IMClientMsg struct {
	Type           string `json:"type"`
	ConversationID uint   `json:"conversationId,omitempty"`
	ContentType    uint8  `json:"contentType,omitempty"`
	Content        string `json:"content,omitempty"`
	ReplyToID      uint   `json:"replyToId,omitempty"`
	ClientMsgID    string `json:"clientMsgId,omitempty"`
}

type IMServerMsg struct {
	Type           string         `json:"type"`
	Content        string         `json:"content,omitempty"`
	ClientMsgID    string         `json:"clientMsgId,omitempty"`
	ID             uint           `json:"id,omitempty"`
	ConversationID uint           `json:"conversationId,omitempty"`
	Data           *IMMessageItem `json:"data,omitempty"`
}
