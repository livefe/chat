package model

type AppRelease struct {
	Base
	Platform    string `gorm:"type:varchar(16);not null;index;comment:平台(android)" json:"platform"`
	Version     string `gorm:"type:varchar(32);not null;comment:版本号(semver)" json:"version"`
	DownloadURL string `gorm:"type:varchar(512);not null;comment:下载地址" json:"downloadUrl"`
	FileSize    int64  `gorm:"type:bigint;not null;default:0;comment:文件大小(字节)" json:"fileSize"`
	Changelog   string `gorm:"type:varchar(1000);not null;default:'';comment:更新说明" json:"changelog"`
	IsCurrent   uint8  `gorm:"type:tinyint unsigned;not null;default:0;index;comment:是否当前对外版本(1是)" json:"isCurrent"`
}

func (AppRelease) TableName() string { return "app_release" }
