package model

type Config struct {
	Base
	Key    string `gorm:"type:varchar(64);not null;uniqueIndex;comment:业务键" json:"key"`
	Value  string `gorm:"type:text;not null;comment:配置 JSON" json:"value"`
	Remark string `gorm:"type:varchar(255);not null;default:'';comment:备注说明" json:"remark"`
}

func (Config) TableName() string { return "config" }
