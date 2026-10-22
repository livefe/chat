package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type UserPunishmentRepository interface {
	Create(ctx context.Context, row *model.UserPunishment) error
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.UserPunishment, int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
}

type userPunishmentRepository struct {
	db *database.DB
}

func NewUserPunishmentRepository(db *database.DB) UserPunishmentRepository {
	return &userPunishmentRepository{db: db}
}

func (r *userPunishmentRepository) Create(ctx context.Context, row *model.UserPunishment) error {
	return database.Conn(ctx, r.db).Create(row).Error
}

func (r *userPunishmentRepository) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.UserPunishment, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.UserPunishment{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.UserPunishment
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *userPunishmentRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.UserPunishment{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}
