package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type IMSupportConfigRepository interface {
	Get(ctx context.Context) (*model.IMSupportConfig, error)
	Save(ctx context.Context, cfg *model.IMSupportConfig) error
}

type imSupportConfigRepository struct {
	db *database.DB
}

func NewIMSupportConfigRepository(db *database.DB) IMSupportConfigRepository {
	return &imSupportConfigRepository{db: db}
}

func (r *imSupportConfigRepository) Get(ctx context.Context) (*model.IMSupportConfig, error) {
	var cfg model.IMSupportConfig
	err := database.Conn(ctx, r.db).Order("id ASC").First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *imSupportConfigRepository) Save(ctx context.Context, cfg *model.IMSupportConfig) error {
	if cfg.ID == 0 {
		return database.Conn(ctx, r.db).Create(cfg).Error
	}
	return database.Conn(ctx, r.db).Save(cfg).Error
}
