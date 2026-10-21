package types

type SubmitReportRequest struct {
	TargetType uint8  `json:"targetType" binding:"required,min=1,max=3"`
	TargetID   uint   `json:"targetId" binding:"required"`
	Reason     uint8  `json:"reason" binding:"required,min=1,max=7"`
	Content    string `json:"content" binding:"max=500"`
	Images     []uint `json:"images"`
}

type AdminReportListRequest struct {
	Status     *uint8 `form:"status"`
	TargetType *uint8 `form:"targetType"`
	Reason     *uint8 `form:"reason"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminReportHandleRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Status       uint8  `json:"status" binding:"required,min=1,max=2"`
	HandleResult string `json:"handleResult"`
	Action       uint8  `json:"action" binding:"max=5"`
	MuteHours    int    `json:"muteHours"`
}

type ReportContextSnapshot struct {
	Nickname       string   `json:"nickname,omitempty"`
	Avatar         string   `json:"avatar,omitempty"`
	PostTitle      string   `json:"postTitle,omitempty"`
	PostContent    string   `json:"postContent,omitempty"`
	PostImages     []string `json:"postImages,omitempty"`
	MessageContent string   `json:"messageContent,omitempty"`
	MessageType    uint8    `json:"messageType,omitempty"`
	Scope          uint8    `json:"scope,omitempty"`
	RoomID         uint     `json:"roomId,omitempty"`
	PeerID         uint     `json:"peerId,omitempty"`
	SentAt         string   `json:"sentAt,omitempty"`
}

type ReportItem struct {
	ID                uint                  `json:"id"`
	ReporterID        uint                  `json:"reporterId"`
	ReporterNickname  string                `json:"reporterNickname,omitempty"`
	ReporterMobile    string                `json:"reporterMobile,omitempty"`
	TargetType        uint8                 `json:"targetType"`
	TargetTypeText    string                `json:"targetTypeText"`
	TargetID          uint                  `json:"targetId"`
	TargetUserID      uint                  `json:"targetUserId"`
	TargetUserName    string                `json:"targetUserName,omitempty"`
	Reason            uint8                 `json:"reason"`
	ReasonText        string                `json:"reasonText"`
	Content           string                `json:"content"`
	Images            []string              `json:"images"`
	Context           ReportContextSnapshot `json:"context"`
	Status            uint8                 `json:"status"`
	StatusText        string                `json:"statusText"`
	HandleResult      string                `json:"handleResult"`
	HandleAction      uint8                 `json:"handleAction"`
	HandleActionText  string                `json:"handleActionText"`
	EmployeeID        uint                  `json:"employeeId"`
	EmployeeName      string                `json:"employeeName,omitempty"`
	HandledAt         string                `json:"handledAt"`
	TargetReportCount int64                 `json:"targetReportCount,omitempty"`
	CreatedAt         string                `json:"createdAt"`
	UpdatedAt         string                `json:"updatedAt"`
}

type ReportListResponse struct {
	Total int64        `json:"total"`
	List  []ReportItem `json:"list"`
}
