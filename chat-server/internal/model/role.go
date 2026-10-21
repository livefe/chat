package model

const (
	RoleStatusDisabled uint8 = 0
	RoleStatusEnabled  uint8 = 1
)

type Role struct {
	Base
	Name        string `gorm:"type:varchar(100);not null;comment:角色名称" json:"name"`
	Code        string `gorm:"type:varchar(50);not null;uniqueIndex;comment:角色编码" json:"code"`
	Description string `gorm:"type:varchar(200);comment:角色描述" json:"description"`
	Status      uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:状态(0:禁用 1:启用)" json:"status"`
}

func (Role) TableName() string { return "role" }
