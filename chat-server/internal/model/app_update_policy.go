package model

type AppUpdatePolicy struct {
	Base
	Platform            string `gorm:"type:varchar(16);not null;uniqueIndex;comment:平台(android/ios/h5)" json:"platform"`
	MinSupportedVersion string `gorm:"type:varchar(32);not null;default:'1.0.0';comment:最低可运行版本" json:"minSupportedVersion"`
	Title               string `gorm:"type:varchar(80);not null;default:'';comment:更新弹窗标题" json:"title"`
	Message             string `gorm:"type:varchar(500);not null;default:'';comment:更新说明" json:"message"`
}

func (AppUpdatePolicy) TableName() string { return "app_update_policy" }
