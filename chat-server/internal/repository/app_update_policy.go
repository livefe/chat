package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type AppUpdatePolicyRepository interface {
	GetByPlatform(ctx context.Context, platform string) (*model.AppUpdatePolicy, error)
	List(ctx context.Context) ([]*model.AppUpdatePolicy, error)
	Upsert(ctx context.Context, row *model.AppUpdatePolicy) error
}

type appUpdatePolicyRepository struct {
	db *database.DB
}

func NewAppUpdatePolicyRepository(db *database.DB) AppUpdatePolicyRepository {
	return &appUpdatePolicyRepository{db: db}
}

func (r *appUpdatePolicyRepository) GetByPlatform(ctx context.Context, platform string) (*model.AppUpdatePolicy, error) {
	var row model.AppUpdatePolicy
	err := database.Conn(ctx, r.db).Where("platform = ?", platform).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *appUpdatePolicyRepository) List(ctx context.Context) ([]*model.AppUpdatePolicy, error) {
	var list []*model.AppUpdatePolicy
	err := database.Conn(ctx, r.db).Order("platform ASC").Find(&list).Error
	return list, err
}

func (r *appUpdatePolicyRepository) Upsert(ctx context.Context, row *model.AppUpdatePolicy) error {
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "platform"}},
		DoUpdates: clause.AssignmentColumns([]string{"min_supported_version", "title", "message", "updated_at"}),
	}).Create(row).Error
}
