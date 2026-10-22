package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type FeedbackListFilter struct {
	UserID    uint
	Status    *uint8
	Type      *uint8
	Keyword   string
	StartTime string
	EndTime   string
	Page      int
	PageSize  int
}

type FeedbackRepository interface {
	Create(ctx context.Context, fb *model.Feedback) error
	Update(ctx context.Context, fb *model.Feedback) error
	GetByID(ctx context.Context, id uint) (*model.Feedback, error)
	List(ctx context.Context, f FeedbackListFilter) ([]*model.Feedback, int64, error)
	CountByUserBetween(ctx context.Context, userID uint, start, end time.Time) (int64, error)
	CountByStatus(ctx context.Context, status uint8) (int64, error)
	UpdatePendingByIDs(ctx context.Context, ids []uint, updates map[string]any) (int64, error)
}

type feedbackRepository struct {
	db *database.DB
}

func NewFeedbackRepository(db *database.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) Create(ctx context.Context, fb *model.Feedback) error {
	return database.Conn(ctx, r.db).Create(fb).Error
}

func (r *feedbackRepository) Update(ctx context.Context, fb *model.Feedback) error {
	return database.Conn(ctx, r.db).Save(fb).Error
}

func (r *feedbackRepository) GetByID(ctx context.Context, id uint) (*model.Feedback, error) {
	var fb model.Feedback
	if err := database.Conn(ctx, r.db).First(&fb, id).Error; err != nil {
		return nil, err
	}
	return &fb, nil
}

func applyFeedbackListFilter(db *gorm.DB, f FeedbackListFilter) *gorm.DB {
	if f.UserID > 0 {
		db = db.Where("user_id = ?", f.UserID)
	}
	if f.Status != nil {
		db = db.Where("`status` = ?", *f.Status)
	}
	if f.Type != nil {
		db = db.Where("`type` = ?", *f.Type)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		db = db.Where("content LIKE ? OR contact LIKE ? OR handle_result LIKE ?", kw, kw, kw)
	}
	if f.StartTime != "" {
		db = db.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime != "" {
		db = db.Where("created_at <= ?", f.EndTime)
	}
	return db
}

func (r *feedbackRepository) UpdatePendingByIDs(ctx context.Context, ids []uint, updates map[string]any) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := database.Conn(ctx, r.db).Model(&model.Feedback{}).
		Where("id IN ? AND `status` = ?", ids, model.FeedbackStatusPending).
		Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *feedbackRepository) List(ctx context.Context, f FeedbackListFilter) ([]*model.Feedback, int64, error) {
	db := applyFeedbackListFilter(database.Conn(ctx, r.db).Model(&model.Feedback{}), f)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.Feedback
	if err := db.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *feedbackRepository) CountByUserBetween(ctx context.Context, userID uint, start, end time.Time) (int64, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.Feedback{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Count(&count).Error
	return count, err
}

func (r *feedbackRepository) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Feedback{}).Where("status = ?", status).Count(&n).Error
	return n, err
}
