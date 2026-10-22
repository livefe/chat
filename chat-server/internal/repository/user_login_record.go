package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"
)

type UserLoginRecordRepository interface {
	Create(ctx context.Context, record *model.UserLoginRecord) error
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.UserLoginRecord, int64, error)
	LastSuccessAt(ctx context.Context, userIDs []uint) (map[uint]time.Time, error)
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

func (r *userLoginRecordRepository) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]*model.UserLoginRecord, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.UserLoginRecord{}).Where("user_id = ?", userID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.UserLoginRecord
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *userLoginRecordRepository) LastSuccessAt(ctx context.Context, userIDs []uint) (map[uint]time.Time, error) {
	out := make(map[uint]time.Time, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		UserID uint
		LastAt time.Time
	}
	err := database.Conn(ctx, r.db).Model(&model.UserLoginRecord{}).
		Select("user_id, MAX(created_at) AS last_at").
		Where("user_id IN ? AND status = 1", userIDs).
		Group("user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.UserID] = row.LastAt
	}
	return out, nil
}
