package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type UserDeactivationLogRepository interface {
	Create(ctx context.Context, log *model.UserDeactivationLog) error
	ListByUserID(ctx context.Context, userID uint, limit int) ([]*model.UserDeactivationLog, error)
}

type userDeactivationLogRepository struct {
	db *database.DB
}

func NewUserDeactivationLogRepository(db *database.DB) UserDeactivationLogRepository {
	return &userDeactivationLogRepository{db: db}
}

func (r *userDeactivationLogRepository) Create(ctx context.Context, log *model.UserDeactivationLog) error {
	return database.Conn(ctx, r.db).Create(log).Error
}

func (r *userDeactivationLogRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]*model.UserDeactivationLog, error) {
	var list []*model.UserDeactivationLog
	err := database.Conn(ctx, r.db).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
