package types

type AuditLogListRequest struct {
	EmployeeID *uint  `form:"employeeId"`
	Module     string `form:"module"`
	Action     string `form:"action"`
	Keyword    string `form:"keyword"`
	StartTime  string `form:"startTime"`
	EndTime    string `form:"endTime"`
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"pageSize" binding:"required,min=1,max=100"`
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
	TargetID       uint   `json:"targetId"`
	TargetLabel    string `json:"targetLabel"`
	RequestPath    string `json:"requestPath"`
	RequestMethod  string `json:"requestMethod"`
	Success        bool   `json:"success"`
	SuccessText    string `json:"successText"`
	Detail         string `json:"detail"`
	IP             string `json:"ip"`
	UserAgent      string `json:"userAgent"`
	CreatedAt      string `json:"createdAt"`
}

type AuditModuleOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type AuditActionOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type AuditLogListResponse struct {
	Total   int64               `json:"total"`
	List    []AuditLogItem      `json:"list"`
	Modules []AuditModuleOption `json:"modules"`
	Actions []AuditActionOption `json:"actions"`
}

type AuditLogDetailResponse = AuditLogItem
