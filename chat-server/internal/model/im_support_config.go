package model

type SupportQuickReply struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type IMSupportConfig struct {
	Base
	BrandName          string              `gorm:"type:varchar(32);not null;default:'官方客服';comment:展示名" json:"brandName"`
	BrandAvatarMediaID uint                `gorm:"type:int unsigned;not null;default:0;comment:展示头像媒体ID" json:"brandAvatarMediaId"`
	WelcomeText        string              `gorm:"type:varchar(500);not null;default:'您好，请描述您的问题，我们会尽快回复';comment:欢迎语" json:"welcomeText"`
	OfflineText        string              `gorm:"type:varchar(500);not null;default:'当前客服繁忙或不在线，留言后我们会尽快回复';comment:离线提示" json:"offlineText"`
	Enabled            uint8               `gorm:"type:tinyint unsigned;not null;default:1;comment:是否启用" json:"enabled"`
	QuickReplies       []SupportQuickReply `gorm:"type:json;serializer:json;comment:快捷回复[{title,content}]" json:"quickReplies"`
}

func (IMSupportConfig) TableName() string { return "im_support_config" }
