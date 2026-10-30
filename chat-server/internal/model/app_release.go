package model

type AppRelease struct {
	Base
	Platform  string `gorm:"type:varchar(16);not null;index;comment:平台(android/ios/h5)" json:"platform"`
	Version   string `gorm:"type:varchar(32);not null;comment:版本号(semver)" json:"version"`
	ObjectKey string `gorm:"type:varchar(255);not null;default:'';comment:安装包对象键(android/ios)" json:"objectKey"`
	EntryURL  string `gorm:"type:varchar(512);not null;default:'';comment:网页版入口地址(h5)" json:"entryUrl"`
	FileSize  int64  `gorm:"type:bigint;not null;default:0;comment:文件大小(字节)" json:"fileSize"`
	Changelog string `gorm:"type:varchar(1000);not null;default:'';comment:更新说明" json:"changelog"`
	IsCurrent uint8  `gorm:"type:tinyint unsigned;not null;default:0;index;comment:是否当前对外版本(1是)" json:"isCurrent"`
}

func (AppRelease) TableName() string { return "app_release" }
