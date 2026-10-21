package model

const (
	DeactivationSourceUser  uint8 = 1
	DeactivationSourceAdmin uint8 = 2
)

type UserDeactivationLog struct {
	Base
	UserID     uint   `gorm:"type:int unsigned;not null;index;comment:用户ID" json:"userId"`
	Mobile     string `gorm:"type:varchar(11);not null;index;comment:原手机号" json:"mobile"`
	Source     uint8  `gorm:"type:tinyint unsigned;not null;index;comment:来源(1:用户自助 2:管理员)" json:"source"`
	EmployeeID uint   `gorm:"type:int unsigned;not null;default:0;comment:操作员工ID" json:"employeeId"`
	Reason     string `gorm:"type:varchar(200);not null;default:'';comment:注销原因" json:"reason"`
	IP         string `gorm:"type:varchar(50);not null;default:'';comment:IP" json:"ip"`
	UserAgent  string `gorm:"type:varchar(500);not null;default:'';comment:UA" json:"userAgent"`
}

func (UserDeactivationLog) TableName() string { return "user_deactivation_log" }
