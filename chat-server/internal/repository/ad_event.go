package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type AdEventRepository interface {
	CreateIgnoreDuplicate(ctx context.Context, item *model.AdEvent) (inserted bool, err error)
}

type adEventRepository struct{ db *database.DB }

func NewAdEventRepository(db *database.DB) AdEventRepository {
	return &adEventRepository{db: db}
}

func (r *adEventRepository) CreateIgnoreDuplicate(ctx context.Context, item *model.AdEvent) (bool, error) {
	res := database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "request_id"}, {Name: "event_type"}},
		DoNothing: true,
	}).Create(item)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
