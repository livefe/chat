package model

import "time"

type ConversationPin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"type:int unsigned;not null;uniqueIndex:uk_user_peer,priority:1;comment:用户ID" json:"userId"`
	PeerID    uint      `gorm:"type:int unsigned;not null;uniqueIndex:uk_user_peer,priority:2;comment:对方用户ID" json:"peerId"`
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;type:datetime" json:"updatedAt"`
}

func (ConversationPin) TableName() string { return "conversation_pin" }
