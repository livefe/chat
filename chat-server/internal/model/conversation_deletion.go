package model

import "time"

type ConversationDeletion struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"type:int unsigned;not null;uniqueIndex:uk_user_peer,priority:1;comment:操作用户ID" json:"userId"`
	PeerID          uint      `gorm:"type:int unsigned;not null;uniqueIndex:uk_user_peer,priority:2;comment:对方用户ID" json:"peerId"`
	ClearedBeforeID uint      `gorm:"column:cleared_before_id;type:int unsigned;not null;default:0;comment:清除时最后一条消息ID，此前消息对该用户不可见" json:"clearedBeforeId"`
	CreatedAt       time.Time `gorm:"autoCreateTime;type:datetime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime;type:datetime" json:"updatedAt"`
}

func (ConversationDeletion) TableName() string { return "conversation_deletion" }
