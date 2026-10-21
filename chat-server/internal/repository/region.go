package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type RegionRepository interface {
	UpsertBatch(ctx context.Context, items []model.Region) (inserted int, updated int, err error)
	RevokeExcept(ctx context.Context, codes []string) (int64, error)
	ListChildren(ctx context.Context, parentCode string) ([]*model.Region, error)
	Search(ctx context.Context, keyword string, level uint8, limit int) ([]*model.Region, error)
	GetByCode(ctx context.Context, code string) (*model.Region, error)
}

type regionRepository struct {
	db *database.DB
}

func NewRegionRepository(db *database.DB) RegionRepository {
	return &regionRepository{db: db}
}

func (r *regionRepository) UpsertBatch(ctx context.Context, items []model.Region) (int, int, error) {
	if len(items) == 0 {
		return 0, 0, nil
	}

	var existing int64
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	if err := database.Conn(ctx, r.db).Model(&model.Region{}).
		Where("code IN ?", codes).
		Count(&existing).Error; err != nil {
		return 0, 0, err
	}

	err := database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "short_name", "parent_code", "level", "full_path",
			"tianditu_gb", "amap_adcode", "data_year", "status", "updated_at",
		}),
	}).CreateInBatches(items, 200).Error
	if err != nil {
		return 0, 0, err
	}

	inserted := len(items) - int(existing)
	if inserted < 0 {
		inserted = 0
	}
	return inserted, int(existing), nil
}

func (r *regionRepository) RevokeExcept(ctx context.Context, codes []string) (int64, error) {
	const batchSize = 500

	var active []string
	if err := database.Conn(ctx, r.db).Model(&model.Region{}).
		Where("status = ?", model.RegionStatusActive).
		Pluck("code", &active).Error; err != nil {
		return 0, err
	}
	if len(active) == 0 {
		return 0, nil
	}

	keep := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		keep[code] = struct{}{}
	}

	revoke := make([]string, 0)
	for _, code := range active {
		if _, ok := keep[code]; !ok {
			revoke = append(revoke, code)
		}
	}
	if len(revoke) == 0 {
		return 0, nil
	}

	var total int64
	for i := 0; i < len(revoke); i += batchSize {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		end := i + batchSize
		if end > len(revoke) {
			end = len(revoke)
		}
		batch := revoke[i:end]
		result := database.Conn(ctx, r.db).Model(&model.Region{}).
			Where("code IN ? AND status = ?", batch, model.RegionStatusActive).
			Update("status", model.RegionStatusRevoked)
		if result.Error != nil {
			return total, result.Error
		}
		total += result.RowsAffected
	}
	return total, nil
}

func (r *regionRepository) ListChildren(ctx context.Context, parentCode string) ([]*model.Region, error) {
	var list []*model.Region
	err := database.Conn(ctx, r.db).
		Where("parent_code = ? AND status = ?", parentCode, model.RegionStatusActive).
		Order("code ASC").
		Find(&list).Error
	return list, err
}

func (r *regionRepository) Search(ctx context.Context, keyword string, level uint8, limit int) ([]*model.Region, error) {
	db := database.Conn(ctx, r.db).Model(&model.Region{}).
		Where("status = ?", model.RegionStatusActive)
	if level > 0 {
		db = db.Where("level = ?", level)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR short_name LIKE ?", like, like)
	}
	var list []*model.Region
	err := db.Order("level ASC, code ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *regionRepository) GetByCode(ctx context.Context, code string) (*model.Region, error) {
	var item model.Region
	if err := database.Conn(ctx, r.db).Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
