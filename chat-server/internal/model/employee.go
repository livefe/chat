package model

const (
	EmployeeStatusDisabled uint8 = 0
	EmployeeStatusEnabled  uint8 = 1
)

type Employee struct {
	Base
	DepartmentID uint   `gorm:"type:int unsigned;not null;default:0;index;comment:部门ID" json:"departmentId"`
	Mobile       string `gorm:"type:varchar(11);not null;uniqueIndex;comment:登录手机号" json:"mobile"`
	Name         string `gorm:"type:varchar(32);not null;default:'';comment:姓名" json:"name"`
	EmployeeNo   string `gorm:"type:varchar(32);not null;default:'';index;comment:工号" json:"employeeNo"`
	AvatarURL    string `gorm:"type:varchar(255);not null;default:'';comment:头像URL" json:"avatarUrl"`
	Status       uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:状态(0:禁用 1:启用)" json:"status"`
	Remark       string `gorm:"type:varchar(255);not null;default:'';comment:备注" json:"remark"`
}

func (Employee) TableName() string { return "employee" }
