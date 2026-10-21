package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"
)

type MediaAssetRepository interface {
	Create(ctx context.Context, asset *model.MediaAsset) error
	GetByID(ctx context.Context, id uint) (*model.MediaAsset, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.MediaAsset, error)
	Update(ctx context.Context, asset *model.MediaAsset) error
	ListOrphans(ctx context.Context, before time.Time, limit int) ([]*model.MediaAsset, error)
	ListDeletedPendingPurge(ctx context.Context, limit int) ([]*model.MediaAsset, error)
}

type mediaAssetRepository struct {
	db *database.DB
}

func NewMediaAssetRepository(db *database.DB) MediaAssetRepository {
	return &mediaAssetRepository{db: db}
}

func (r *mediaAssetRepository) Create(ctx context.Context, asset *model.MediaAsset) error {
	return database.Conn(ctx, r.db).Create(asset).Error
}

func (r *mediaAssetRepository) GetByID(ctx context.Context, id uint) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := database.Conn(ctx, r.db).First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *mediaAssetRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.MediaAsset, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*model.MediaAsset
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *mediaAssetRepository) Update(ctx context.Context, asset *model.MediaAsset) error {
	return database.Conn(ctx, r.db).Save(asset).Error
}

func (r *mediaAssetRepository) ListOrphans(ctx context.Context, before time.Time, limit int) ([]*model.MediaAsset, error) {
	var list []*model.MediaAsset
	err := database.Conn(ctx, r.db).
		Where("status = ? AND created_at < ?", model.MediaStatusUploaded, before).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *mediaAssetRepository) ListDeletedPendingPurge(ctx context.Context, limit int) ([]*model.MediaAsset, error) {
	var list []*model.MediaAsset
	err := database.Conn(ctx, r.db).
		Where("status = ? AND object_key <> ''", model.MediaStatusDeleted).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
