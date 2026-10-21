package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type ConfigRepository interface {
	GetByKey(ctx context.Context, key string) (*model.Config, error)
	Upsert(ctx context.Context, key, value, remark string) error
}

type configRepository struct {
	db *database.DB
}

func NewConfigRepository(db *database.DB) ConfigRepository {
	return &configRepository{db: db}
}

func (r *configRepository) GetByKey(ctx context.Context, key string) (*model.Config, error) {
	var item model.Config
	if err := database.Conn(ctx, r.db).Where("`key` = ?", key).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *configRepository) Upsert(ctx context.Context, key, value, remark string) error {
	row := &model.Config{Key: key, Value: value, Remark: remark}
	updates := map[string]any{
		"value":      value,
		"updated_at": time.Now(),
	}
	if remark != "" {
		updates["remark"] = remark
	}
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.Assignments(updates),
	}).Create(row).Error
}
