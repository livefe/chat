package model

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;type:datetime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;type:datetime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:datetime" json:"-"`
}
