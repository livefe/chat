package types

const (
	ScopeSupport = "support"

	CacheSupportUnreadKey = cachePrefix + "support:unread:%d"

	SupportCloseReasonResolved   = "resolved"
	SupportCloseReasonNoResponse = "no_response"
	SupportCloseReasonDuplicate  = "duplicate"
	SupportCloseReasonOther      = "other"
	SupportCloseReasonUserCancel = "user_cancel"

	SupportQuickReplyMax = 50
)

type SupportProfileResponse struct {
	BrandName      string `json:"brandName"`
	BrandAvatarURL string `json:"brandAvatarUrl"`
	WelcomeText    string `json:"welcomeText"`
	OfflineText    string `json:"offlineText"`
	Enabled        bool   `json:"enabled"`
}

type SupportQuickReplyItem struct {
	Title   string `json:"title" binding:"required,max=20"`
	Content string `json:"content" binding:"required,max=500"`
}

type SupportAdminConfigResponse struct {
	SupportProfileResponse
	QuickReplies []SupportQuickReplyItem `json:"quickReplies"`
}

type SupportQuickReplyListResponse struct {
	List []SupportQuickReplyItem `json:"list"`
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
	CloseReason        string `json:"closeReason"`
	CloseReasonText    string `json:"closeReasonText"`
	CloseNote          string `json:"closeNote"`
}

type SupportAdminSessionListResponse struct {
	Total int64                     `json:"total"`
	List  []SupportAdminSessionItem `json:"list"`
}

type SupportQueueQuery struct {
	PageQuery
	Keyword string `form:"keyword" binding:"max=64"`
}

type SupportOverviewQuery struct {
	PageQuery
	AssigneeEmployeeID uint   `form:"assigneeEmployeeId"`
	Keyword            string `form:"keyword" binding:"max=64"`
	StartDate          string `form:"startDate" binding:"omitempty,datetime=2006-01-02"`
	EndDate            string `form:"endDate" binding:"omitempty,datetime=2006-01-02"`
}

type SupportSessionQuery struct {
	ConversationID uint `form:"conversationId" binding:"required"`
}

type SupportByUserQuery struct {
	PageQuery
	UserID uint `form:"userId" binding:"required"`
}

type SupportHistoryQuery struct {
	PageQuery
	ConversationID uint `form:"conversationId" binding:"required"`
}

type SupportUserSessionItem struct {
	ID              uint   `json:"id"`
	Status          uint8  `json:"status"`
	StatusText      string `json:"statusText"`
	AssigneeName    string `json:"assigneeName"`
	CreatedAt       string `json:"createdAt"`
	AcceptedAt      string `json:"acceptedAt"`
	ClosedAt        string `json:"closedAt"`
	CloseReason     string `json:"closeReason"`
	CloseReasonText string `json:"closeReasonText"`
	CloseNote       string `json:"closeNote"`
	LastMessage     string `json:"lastMessage"`
	LastMessageAt   string `json:"lastMessageAt"`
}

type SupportUserSessionListResponse struct {
	Total int64                    `json:"total"`
	List  []SupportUserSessionItem `json:"list"`
}

type SupportAgentItem struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	AvatarURL      string `json:"avatarUrl"`
	DepartmentName string `json:"departmentName"`
	Online         bool   `json:"online"`
}

type SupportAgentListResponse struct {
	List []SupportAgentItem `json:"list"`
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
	Reason         string `json:"reason" binding:"omitempty,oneof=resolved no_response duplicate other"`
	Note           string `json:"note" binding:"max=500"`
}

type SupportConfigUpdateRequest struct {
	BrandName      string                  `json:"brandName" binding:"max=32"`
	BrandAvatarURL string                  `json:"brandAvatarUrl" binding:"max=255"`
	WelcomeText    string                  `json:"welcomeText" binding:"max=500"`
	OfflineText    string                  `json:"offlineText" binding:"max=500"`
	Enabled        *uint8                  `json:"enabled" binding:"omitempty,oneof=0 1"`
	QuickReplies   []SupportQuickReplyItem `json:"quickReplies" binding:"max=50,dive"`
}
