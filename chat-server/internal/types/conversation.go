package types

type ConversationItem struct {
	PeerID       uint   `json:"peerId"`
	PeerNickname string `json:"peerNickname"`
	PeerAvatar   string `json:"peerAvatar"`
	LastContent  string `json:"lastContent"`
	LastActiveAt int64  `json:"lastActiveAt"`
	UnreadCount  int64  `json:"unreadCount"`
	Pinned       bool   `json:"pinned"`
}

type ConversationListResponse struct {
	Total int64              `json:"total"`
	List  []ConversationItem `json:"list"`
}

type ConversationListQuery struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type ConversationUnreadSummary struct {
	Total int64                    `json:"total"`
	Items []ConversationUnreadItem `json:"items"`
}

type ConversationUnreadItem struct {
	PeerID uint  `json:"peerId"`
	Count  int64 `json:"count"`
}

type PeerIDRequest struct {
	PeerID uint `json:"peerId" binding:"required"`
}

type PinConversationRequest struct {
	PeerID uint `json:"peerId" binding:"required"`
	Pinned bool `json:"pinned"`
}

type DeleteConversationResponse struct {
	ClearedBeforeID uint `json:"clearedBeforeId"`
}

type AdminConversationListRequest struct {
	UserID     uint   `form:"userId"`
	PeerUserID uint   `form:"peerUserId"`
	StartTime  string `form:"startTime"`
	EndTime    string `form:"endTime"`
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminConversationItem struct {
	ID            uint   `json:"id"`
	UserMinID     uint   `json:"userMinId"`
	UserMaxID     uint   `json:"userMaxId"`
	UserMinName   string `json:"userMinName"`
	UserMaxName   string `json:"userMaxName"`
	UserMinMobile string `json:"userMinMobile"`
	UserMaxMobile string `json:"userMaxMobile"`
	LastContent   string `json:"lastContent"`
	LastActiveAt  int64  `json:"lastActiveAt"`
	CreatedAt     string `json:"createdAt"`
}

type AdminConversationListResponse struct {
	Total int64                   `json:"total"`
	List  []AdminConversationItem `json:"list"`
}
