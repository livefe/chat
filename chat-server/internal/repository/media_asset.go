package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

const mediaUnusedCond = "NOT EXISTS (SELECT 1 FROM media_usage WHERE media_usage.media_id = media_asset.id)"

type MediaAssetRepository interface {
	Create(ctx context.Context, asset *model.MediaAsset) error
	GetByID(ctx context.Context, id uint) (*model.MediaAsset, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.MediaAsset, error)
	MarkUploaded(ctx context.Context, id uint, size int64) (bool, error)
	Discard(ctx context.Context, id uint) error
	PromoteBound(ctx context.Context, id uint, at time.Time) error
	LockBound(ctx context.Context, id uint) (bool, error)
	MarkDeletedIfUnused(ctx context.Context, id uint) error
	ListOrphanIDs(ctx context.Context, before time.Time, afterID uint, limit int) ([]uint, error)
	MarkOrphanDeleted(ctx context.Context, id uint, before time.Time) (bool, error)
	ListPurgeable(ctx context.Context, afterID uint, limit int) ([]*model.MediaAsset, error)
	MarkPurged(ctx context.Context, id uint, at time.Time) error
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

func (r *mediaAssetRepository) MarkUploaded(ctx context.Context, id uint, size int64) (bool, error) {
	res := database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("id = ? AND status = ?", id, model.MediaStatusPending).
		Updates(map[string]any{"status": model.MediaStatusUploaded, "size_bytes": size})
	return res.RowsAffected > 0, res.Error
}

func (r *mediaAssetRepository) Discard(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("id = ? AND status = ?", id, model.MediaStatusPending).
		Update("status", model.MediaStatusDeleted).Error
}

func (r *mediaAssetRepository) PromoteBound(ctx context.Context, id uint, at time.Time) error {
	return database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("id = ? AND status = ?", id, model.MediaStatusUploaded).
		Updates(map[string]any{"status": model.MediaStatusBound, "bound_at": at}).Error
}

func (r *mediaAssetRepository) LockBound(ctx context.Context, id uint) (bool, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND status = ?", id, model.MediaStatusBound).
		Count(&n).Error
	return n > 0, err
}

func (r *mediaAssetRepository) MarkDeletedIfUnused(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("id = ? AND status IN ?", id, []uint8{model.MediaStatusUploaded, model.MediaStatusBound}).
		Where(mediaUnusedCond).
		Update("status", model.MediaStatusDeleted).Error
}

func (r *mediaAssetRepository) ListOrphanIDs(ctx context.Context, before time.Time, afterID uint, limit int) ([]uint, error) {
	var ids []uint
	err := database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("status IN ? AND created_at < ? AND id > ?", []uint8{model.MediaStatusPending, model.MediaStatusUploaded}, before, afterID).
		Order("id ASC").
		Limit(limit).
		Pluck("id", &ids).Error
	return ids, err
}

func (r *mediaAssetRepository) MarkOrphanDeleted(ctx context.Context, id uint, before time.Time) (bool, error) {
	res := database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("id = ? AND status IN ? AND created_at < ?", id, []uint8{model.MediaStatusPending, model.MediaStatusUploaded}, before).
		Where(mediaUnusedCond).
		Update("status", model.MediaStatusDeleted)
	return res.RowsAffected > 0, res.Error
}

func (r *mediaAssetRepository) ListPurgeable(ctx context.Context, afterID uint, limit int) ([]*model.MediaAsset, error) {
	var list []*model.MediaAsset
	err := database.Conn(ctx, r.db).
		Where("status = ? AND purged_at IS NULL AND id > ?", model.MediaStatusDeleted, afterID).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *mediaAssetRepository) MarkPurged(ctx context.Context, id uint, at time.Time) error {
	return database.Conn(ctx, r.db).Model(&model.MediaAsset{}).
		Where("id = ? AND status = ?", id, model.MediaStatusDeleted).
		Update("purged_at", at).Error
}
