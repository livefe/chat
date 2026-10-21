package model

type Conversation struct {
	Base
	UserMinID     uint   `gorm:"type:int unsigned;not null;uniqueIndex:uk_pair,priority:1;comment:较小用户ID" json:"userMinId"`
	UserMaxID     uint   `gorm:"type:int unsigned;not null;uniqueIndex:uk_pair,priority:2;comment:较大用户ID" json:"userMaxId"`
	LastMessageID uint   `gorm:"type:int unsigned;not null;default:0;comment:最后一条消息ID" json:"lastMessageId"`
	LastContent   string `gorm:"type:varchar(200);not null;default:'';comment:最后一条内容预览" json:"lastContent"`
	LastActiveAt  int64  `gorm:"type:bigint;not null;default:0;index;comment:最后活跃时间(秒)" json:"lastActiveAt"`
}

func (Conversation) TableName() string { return "conversation" }
