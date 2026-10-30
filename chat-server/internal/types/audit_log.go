package types

type AuditLogListRequest struct {
	EmployeeID  *uint  `form:"employeeId"`
	Module      string `form:"module"`
	Action      string `form:"action"`
	TargetType  string `form:"targetType"`
	TargetID    *uint  `form:"targetId"`
	RequestPath string `form:"requestPath"`
	Success     *bool  `form:"success"`
	Keyword     string `form:"keyword"`
	StartTime   string `form:"startTime"`
	EndTime     string `form:"endTime"`
	Page        int    `form:"page" binding:"required,min=1"`
	PageSize    int    `form:"pageSize" binding:"required,min=1,max=100"`
}

type AuditLogItem struct {
	ID             uint   `json:"id"`
	EmployeeID     uint   `json:"employeeId"`
	EmployeeName   string `json:"employeeName"`
	EmployeeMobile string `json:"employeeMobile"`
	Module         string `json:"module"`
	ModuleText     string `json:"moduleText"`
	Action         string `json:"action"`
	ActionText     string `json:"actionText"`
	TargetType     string `json:"targetType"`
	TargetTypeText string `json:"targetTypeText"`
	TargetID       uint   `json:"targetId"`
	TargetLabel    string `json:"targetLabel"`
	RequestPath    string `json:"requestPath"`
	RequestMethod  string `json:"requestMethod"`
	Success        bool   `json:"success"`
	SuccessText    string `json:"successText"`
	ResultCode     int    `json:"resultCode"`
	ResultMessage  string `json:"resultMessage"`
	Detail         string `json:"detail"`
	IP             string `json:"ip"`
	UserAgent      string `json:"userAgent"`
	CreatedAt      string `json:"createdAt"`
}

type AuditLogListResponse struct {
	Total int64          `json:"total"`
	List  []AuditLogItem `json:"list"`
}

type AuditLogDetailResponse = AuditLogItem

type AuditOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type AuditActionOption struct {
	Value  string `json:"value"`
	Label  string `json:"label"`
	Module string `json:"module"`
}

type AuditOptionsResponse struct {
	Modules     []AuditOption       `json:"modules"`
	Actions     []AuditActionOption `json:"actions"`
	TargetTypes []AuditOption       `json:"targetTypes"`
}

type AuditEmployeeOptionsRequest struct {
	Keyword string `form:"keyword"`
	ID      uint   `form:"id"`
}

type AuditEmployeeOption struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
}
