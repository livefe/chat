package model

const (
	RegionLevelProvince uint8 = 1
	RegionLevelCity     uint8 = 2
	RegionLevelDistrict uint8 = 3
	RegionLevelTownship uint8 = 4

	RegionStatusRevoked uint8 = 0
	RegionStatusActive  uint8 = 1
)

type Region struct {
	Base
	Code       string `gorm:"type:char(12);not null;uniqueIndex;comment:民政部12位区划码" json:"code"`
	Name       string `gorm:"type:varchar(64);not null;index;comment:标准名称" json:"name"`
	ShortName  string `gorm:"type:varchar(32);not null;default:'';index;comment:简称" json:"shortName"`
	ParentCode string `gorm:"type:char(12);not null;default:'';index;comment:父级区划码" json:"parentCode"`
	Level      uint8  `gorm:"type:tinyint unsigned;not null;index;comment:层级1-4" json:"level"`
	FullPath   string `gorm:"type:varchar(256);not null;default:'';comment:全路径" json:"fullPath"`
	TiandituGB string `gorm:"type:char(9);not null;default:'';index;comment:天地图国标码" json:"tiandituGb"`
	AmapAdcode string `gorm:"type:char(6);not null;default:'';index;comment:高德adcode" json:"amapAdcode"`
	DataYear   string `gorm:"type:char(4);not null;default:'';comment:数据年份" json:"dataYear"`
	Status     uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:1有效0撤销" json:"status"`
}

func (Region) TableName() string { return "region" }
