package types

type SubmitFeedbackRequest struct {
	Type    uint8  `json:"type" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"required,min=10,max=500"`
	Contact string `json:"contact"`
	Images  []uint `json:"images"`
}

type MyFeedbackListRequest struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminFeedbackListRequest struct {
	Status    *uint8 `form:"status"`
	Type      *uint8 `form:"type"`
	UserID    uint   `form:"userId"`
	Keyword   string `form:"keyword"`
	StartTime string `form:"startTime"`
	EndTime   string `form:"endTime"`
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AdminFeedbackUpdateStatusRequest struct {
	ID           uint   `json:"id" binding:"required"`
	Status       uint8  `json:"status" binding:"required,min=1,max=2"`
	HandleResult string `json:"handleResult" binding:"max=500"`
}

type AdminFeedbackBatchCloseRequest struct {
	IDs          []uint `json:"ids" binding:"required,min=1,max=100"`
	HandleResult string `json:"handleResult" binding:"max=500"`
}

type AdminFeedbackBatchResult struct {
	Updated int64 `json:"updated"`
	Skipped int64 `json:"skipped"`
}

type FeedbackItem struct {
	ID           uint     `json:"id"`
	UserID       uint     `json:"userId"`
	UserNickname string   `json:"userNickname,omitempty"`
	UserMobile   string   `json:"userMobile,omitempty"`
	Type         uint8    `json:"type"`
	TypeText     string   `json:"typeText"`
	Content      string   `json:"content"`
	Contact      string   `json:"contact"`
	Images       []string `json:"images"`
	Status       uint8    `json:"status"`
	StatusText   string   `json:"statusText"`
	HandleResult string   `json:"handleResult"`
	EmployeeID   uint     `json:"employeeId"`
	EmployeeName string   `json:"employeeName,omitempty"`
	HandledAt    string   `json:"handledAt"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

type FeedbackListResponse struct {
	Total int64          `json:"total"`
	List  []FeedbackItem `json:"list"`
}
