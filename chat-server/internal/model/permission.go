package model

const (
	PermissionTypeMenu   uint8 = 1
	PermissionTypeButton uint8 = 2
	PermissionTypeAPI    uint8 = 3
	PermissionTypeData   uint8 = 4

	PermissionStatusDisabled uint8 = 0
	PermissionStatusEnabled  uint8 = 1
)

type Permission struct {
	Base
	ParentID    uint   `gorm:"type:int unsigned;default:0;index;comment:父级权限ID" json:"parentId"`
	Name        string `gorm:"type:varchar(100);not null;comment:权限名称" json:"name"`
	Code        string `gorm:"type:varchar(64);not null;index;comment:权限编码(语义码，按钮与接口可同码)" json:"code"`
	Description string `gorm:"type:varchar(200);comment:权限描述" json:"description"`
	Type        uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:权限类型(1:菜单 2:按钮 3:接口 4:数据)" json:"type"`
	Path        string `gorm:"type:varchar(200);comment:权限路径" json:"path"`
	Method      string `gorm:"type:varchar(10);comment:请求方法" json:"method"`
	Icon        string `gorm:"type:varchar(100);comment:菜单图标" json:"icon"`
	Component   string `gorm:"type:varchar(200);comment:前端组件名称" json:"component"`
	HideInMenu  bool   `gorm:"type:boolean;default:false;comment:是否隐藏菜单" json:"hideInMenu"`
	Sort        uint   `gorm:"type:int unsigned;default:0;comment:排序" json:"sort"`
	Status      uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:状态(0:禁用 1:启用)" json:"status"`
}

func (Permission) TableName() string { return "permission" }
