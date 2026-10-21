package model

type Device struct {
	Base
	UserID       uint   `gorm:"type:int unsigned;not null;uniqueIndex:uk_device,priority:1;comment:用户ID" json:"userId"`
	DeviceID     string `gorm:"type:varchar(64);not null;uniqueIndex:uk_device,priority:2;comment:客户端设备ID" json:"deviceId"`
	Platform     string `gorm:"type:varchar(16);not null;default:android;comment:平台" json:"platform"`
	Vendor       string `gorm:"type:varchar(16);not null;index;comment:厂商通道" json:"vendor"`
	PushToken    string `gorm:"type:varchar(512);not null;index;comment:厂商推送token" json:"pushToken"`
	AppVersion   string `gorm:"type:varchar(32);not null;default:'';comment:App版本" json:"appVersion"`
	OSVersion    string `gorm:"type:varchar(32);not null;default:'';comment:系统版本" json:"osVersion"`
	PushEnabled  bool   `gorm:"type:tinyint(1);not null;default:1;comment:是否接收推送" json:"pushEnabled"`
	LastActiveAt int64  `gorm:"type:bigint;not null;default:0;comment:最后活跃时间(秒)" json:"lastActiveAt"`
}

func (Device) TableName() string { return "device" }
