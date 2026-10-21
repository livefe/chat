package model

const (
	DepartmentStatusDisabled uint8 = 0
	DepartmentStatusEnabled  uint8 = 1
)

type Department struct {
	Base
	ParentID uint   `gorm:"type:int unsigned;not null;default:0;index;comment:上级部门ID(0为顶级)" json:"parentId"`
	Name     string `gorm:"type:varchar(64);not null;comment:部门名称" json:"name"`
	Path     string `gorm:"type:varchar(255);not null;default:'';index;comment:路径如/1/3/5/" json:"path"`
	Level    uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:层级(从1开始)" json:"level"`
	Status   uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:状态(0:禁用 1:启用)" json:"status"`
	Sort     int    `gorm:"type:int;not null;default:0;comment:同级排序" json:"sort"`
	Remark   string `gorm:"type:varchar(255);not null;default:'';comment:备注" json:"remark"`
}

func (Department) TableName() string { return "department" }
