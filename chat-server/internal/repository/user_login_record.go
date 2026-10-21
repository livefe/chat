package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type UserLoginRecordRepository interface {
	Create(ctx context.Context, record *model.UserLoginRecord) error
	ListByUserID(ctx context.Context, userID uint, limit int) ([]*model.UserLoginRecord, error)
}

type userLoginRecordRepository struct {
	db *database.DB
}

func NewUserLoginRecordRepository(db *database.DB) UserLoginRecordRepository {
	return &userLoginRecordRepository{db: db}
}

func (r *userLoginRecordRepository) Create(ctx context.Context, record *model.UserLoginRecord) error {
	return database.Conn(ctx, r.db).Create(record).Error
}

func (r *userLoginRecordRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]*model.UserLoginRecord, error) {
	var list []*model.UserLoginRecord
	err := database.Conn(ctx, r.db).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
