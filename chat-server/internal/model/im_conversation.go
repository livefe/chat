package model

const (
	IMConvTypeStaffDM uint8 = 1
	IMConvTypeSupport uint8 = 3

	IMConvStatusActive uint8 = 1
	IMConvStatusClosed uint8 = 2
	IMConvStatusQueued uint8 = 3
)

type IMConversation struct {
	Base
	Type          uint8  `gorm:"type:tinyint unsigned;not null;index;comment:类型(1:staff_dm 3:support)" json:"type"`
	Status        uint8  `gorm:"type:tinyint unsigned;not null;default:1;index;comment:状态(1:active 2:closed)" json:"status"`
	Title         string `gorm:"type:varchar(64);not null;default:'';comment:标题(群名等)" json:"title"`
	LastMessageID uint   `gorm:"type:int unsigned;not null;default:0;comment:最后消息ID" json:"lastMessageId"`
	LastContent   string `gorm:"type:varchar(200);not null;default:'';comment:最后内容预览" json:"lastContent"`
	LastActiveAt  int64  `gorm:"type:bigint;not null;default:0;index;comment:最后活跃时间(秒)" json:"lastActiveAt"`
}

func (IMConversation) TableName() string { return "im_conversation" }
