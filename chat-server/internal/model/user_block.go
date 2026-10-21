package model

import "time"

type UserBlock struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BlockerID uint      `gorm:"type:int unsigned;not null;uniqueIndex:uk_blocker_blocked,priority:1;index:idx_blocked;comment:拉黑人" json:"blockerId"`
	BlockedID uint      `gorm:"type:int unsigned;not null;uniqueIndex:uk_blocker_blocked,priority:2;index:idx_blocked;comment:被拉黑人" json:"blockedId"`
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime" json:"createdAt"`
}

func (UserBlock) TableName() string { return "user_block" }
