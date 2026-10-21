package model

const (
	UserStatusDisabled uint8 = 0
	UserStatusEnabled  uint8 = 1

	DeactivatedNickname = "已注销用户"

	PermanentMuteUntil int64 = -1
)

type User struct {
	Base
	Mobile        string `gorm:"type:varchar(11);not null;uniqueIndex;comment:手机号" json:"mobile"`
	Nickname      string `gorm:"type:varchar(32);not null;default:'';comment:昵称" json:"nickname"`
	AvatarMediaID uint   `gorm:"type:int unsigned;not null;default:0;comment:头像媒体ID" json:"avatarMediaId"`
	Status        uint8  `gorm:"type:tinyint unsigned;not null;default:1;comment:状态(0:禁用 1:启用)" json:"status"`
	MutedUntil    int64  `gorm:"type:bigint;not null;default:0;comment:禁言截止时间(秒),0未禁言,-1永久" json:"mutedUntil"`
}

func (User) TableName() string { return "user" }
