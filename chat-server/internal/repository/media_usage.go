package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type MediaUsageRepository interface {
	Create(ctx context.Context, usage *model.MediaUsage) error
	Delete(ctx context.Context, mediaID uint, bizType string, bizID uint) error
	ListMediaIDsByBiz(ctx context.Context, bizType string, bizID uint) ([]uint, error)
}

type mediaUsageRepository struct {
	db *database.DB
}

func NewMediaUsageRepository(db *database.DB) MediaUsageRepository {
	return &mediaUsageRepository{db: db}
}

func (r *mediaUsageRepository) Create(ctx context.Context, usage *model.MediaUsage) error {
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).Create(usage).Error
}

func (r *mediaUsageRepository) Delete(ctx context.Context, mediaID uint, bizType string, bizID uint) error {
	return database.Conn(ctx, r.db).
		Where("media_id = ? AND biz_type = ? AND biz_id = ?", mediaID, bizType, bizID).
		Delete(&model.MediaUsage{}).Error
}

func (r *mediaUsageRepository) ListMediaIDsByBiz(ctx context.Context, bizType string, bizID uint) ([]uint, error) {
	var ids []uint
	err := database.Conn(ctx, r.db).Model(&model.MediaUsage{}).
		Where("biz_type = ? AND biz_id = ?", bizType, bizID).
		Pluck("media_id", &ids).Error
	return ids, err
}
