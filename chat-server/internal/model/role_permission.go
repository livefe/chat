package model

type RolePermission struct {
	Base
	RoleID       uint `gorm:"type:int unsigned;not null;index;comment:角色ID" json:"roleId"`
	PermissionID uint `gorm:"type:int unsigned;not null;index;comment:权限ID" json:"permissionId"`
}

func (RolePermission) TableName() string { return "role_permission" }
