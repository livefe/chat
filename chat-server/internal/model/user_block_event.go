package model

import "time"

const (
	UserBlockEventActionBlock   uint8 = 1
	UserBlockEventActionUnblock uint8 = 2
)

type UserBlockEvent struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BlockerID uint      `gorm:"type:int unsigned;not null;index:idx_block_event_pair,priority:1;index:idx_block_event_blocker;comment:拉黑人" json:"blockerId"`
	BlockedID uint      `gorm:"type:int unsigned;not null;index:idx_block_event_pair,priority:2;index:idx_block_event_blocked;comment:被拉黑人" json:"blockedId"`
	Action    uint8     `gorm:"type:tinyint unsigned;not null;comment:动作(1:拉黑 2:解除)" json:"action"`
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime;index;comment:发生时间" json:"createdAt"`
}

func (UserBlockEvent) TableName() string { return "user_block_event" }
