package model

type AuditLog struct {
	Base
	EmployeeID     uint   `gorm:"type:int unsigned;not null;index;comment:操作员工ID" json:"employeeId"`
	EmployeeName   string `gorm:"type:varchar(64);not null;default:'';comment:员工姓名" json:"employeeName"`
	EmployeeMobile string `gorm:"type:varchar(11);not null;default:'';comment:员工手机号" json:"employeeMobile"`
	Module         string `gorm:"type:varchar(32);not null;index;comment:模块标识" json:"module"`
	ModuleText     string `gorm:"type:varchar(64);not null;default:'';comment:模块名称" json:"moduleText"`
	Action         string `gorm:"type:varchar(32);not null;index;comment:动作标识" json:"action"`
	ActionText     string `gorm:"type:varchar(64);not null;default:'';comment:动作名称" json:"actionText"`
	TargetType     string `gorm:"type:varchar(32);not null;default:'';comment:目标类型" json:"targetType"`
	TargetID       uint   `gorm:"type:int unsigned;not null;default:0;index;comment:目标ID" json:"targetId"`
	TargetLabel    string `gorm:"type:varchar(128);not null;default:'';comment:目标描述" json:"targetLabel"`
	RequestPath    string `gorm:"type:varchar(128);not null;default:'';comment:请求路径" json:"requestPath"`
	RequestMethod  string `gorm:"type:varchar(10);not null;default:'';comment:请求方法" json:"requestMethod"`
	Success        bool   `gorm:"type:tinyint(1);not null;default:1;index;comment:是否成功" json:"success"`
	Detail         string `gorm:"type:text;comment:请求详情" json:"detail"`
	IP             string `gorm:"type:varchar(50);not null;default:'';comment:IP" json:"ip"`
	UserAgent      string `gorm:"type:varchar(500);not null;default:'';comment:UA" json:"userAgent"`
}

func (AuditLog) TableName() string { return "audit_log" }
