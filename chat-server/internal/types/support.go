package types

const (
	ScopeSupport = "support"

	CacheSupportUnreadKey = cachePrefix + "support:unread:%d"
)

type SupportProfileResponse struct {
	BrandName      string `json:"brandName"`
	BrandAvatarURL string `json:"brandAvatarUrl"`
	WelcomeText    string `json:"welcomeText"`
	OfflineText    string `json:"offlineText"`
	Enabled        bool   `json:"enabled"`
}

type SupportSessionItem struct {
	ID                 uint   `json:"id"`
	Status             uint8  `json:"status"`
	StatusText         string `json:"statusText"`
	LastMessageID      uint   `json:"lastMessageId"`
	LastContent        string `json:"lastContent"`
	LastActiveAt       int64  `json:"lastActiveAt"`
	UnreadCount        int64  `json:"unreadCount"`
	AssigneeEmployeeID uint   `json:"assigneeEmployeeId,omitempty"`
	AssigneeName       string `json:"assigneeName,omitempty"`
	QueueAt            int64  `json:"queueAt,omitempty"`
	BrandName          string `json:"brandName"`
	BrandAvatarURL     string `json:"brandAvatarUrl"`
}

type SupportOpenRequest struct {
	Source string `json:"source"`
}

type SupportMessageQuery struct {
	ConversationID uint `form:"conversationId" binding:"required"`
	BeforeID       uint `form:"beforeId"`
	Limit          int  `form:"limit" binding:"required,min=1,max=100"`
}

type SupportSendRequest struct {
	ConversationID uint   `json:"conversationId" binding:"required"`
	ContentType    uint8  `json:"contentType"`
	Content        string `json:"content" binding:"required"`
	ReplyToID      uint   `json:"replyToId"`
	ClientMsgID    string `json:"clientMsgId"`
}

type SupportReadRequest struct {
	ConversationID uint `json:"conversationId" binding:"required"`
	LastReadMsgID  uint `json:"lastReadMsgId"`
}

type SupportAdminSessionItem struct {
	ID                 uint   `json:"id"`
	Status             uint8  `json:"status"`
	StatusText         string `json:"statusText"`
	UserID             uint   `json:"userId"`
	UserNickname       string `json:"userNickname"`
	UserAvatarURL      string `json:"userAvatarUrl"`
	UserMobile         string `json:"userMobile"`
	AssigneeEmployeeID uint   `json:"assigneeEmployeeId"`
	AssigneeName       string `json:"assigneeName"`
	LastMessageID      uint   `json:"lastMessageId"`
	LastContent        string `json:"lastContent"`
	LastActiveAt       int64  `json:"lastActiveAt"`
	UnreadCount        int64  `json:"unreadCount"`
	QueueAt            int64  `json:"queueAt"`
	AcceptAt           int64  `json:"acceptAt"`
	CloseAt            int64  `json:"closeAt,omitempty"`
	QueueWaitSec       int64  `json:"queueWaitSec"`
}

type SupportAdminSessionListResponse struct {
	Total int64                     `json:"total"`
	List  []SupportAdminSessionItem `json:"list"`
}

type SupportOverviewQuery struct {
	PageQuery
	AssigneeEmployeeID uint   `form:"assigneeEmployeeId"`
	Keyword            string `form:"keyword"`
}

type SupportOverviewSummaryResponse struct {
	QueueCount         int64 `json:"queueCount"`
	ActiveCount        int64 `json:"activeCount"`
	LongWaitQueueCount int64 `json:"longWaitQueueCount"`
}

type SupportAcceptRequest struct {
	ConversationID uint `json:"conversationId" binding:"required"`
}

type SupportTransferRequest struct {
	ConversationID uint `json:"conversationId" binding:"required"`
	ToEmployeeID   uint `json:"toEmployeeId" binding:"required"`
}

type SupportCloseRequest struct {
	ConversationID uint   `json:"conversationId" binding:"required"`
	Reason         string `json:"reason"`
}

type SupportConfigUpdateRequest struct {
	BrandName      string `json:"brandName"`
	BrandAvatarURL string `json:"brandAvatarUrl"`
	WelcomeText    string `json:"welcomeText"`
	OfflineText    string `json:"offlineText"`
	Enabled        *uint8 `json:"enabled"`
}
