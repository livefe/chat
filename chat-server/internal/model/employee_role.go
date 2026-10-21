package model

type EmployeeRole struct {
	Base
	EmployeeID uint `gorm:"type:int unsigned;not null;index;comment:员工ID" json:"employeeId"`
	RoleID     uint `gorm:"type:int unsigned;not null;index;comment:角色ID" json:"roleId"`
}

func (EmployeeRole) TableName() string { return "employee_role" }
