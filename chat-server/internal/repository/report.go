package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type ReportListFilter struct {
	ReporterID *uint
	Status     *uint8
	TargetType *uint8
	Reason     *uint8
	Keyword    string
	Page       int
	PageSize   int
}

type ReportRepository interface {
	Create(ctx context.Context, r *model.Report) error
	Update(ctx context.Context, r *model.Report) error
	GetByID(ctx context.Context, id uint) (*model.Report, error)
	List(ctx context.Context, f ReportListFilter) ([]*model.Report, int64, error)
	CountByReporterTargetSince(ctx context.Context, reporterID uint, targetType uint8, targetID uint, since time.Time) (int64, error)
	CountByReporterBetween(ctx context.Context, reporterID uint, start, end time.Time) (int64, error)
	CountByTarget(ctx context.Context, targetType uint8, targetID uint) (int64, error)
	CountByStatus(ctx context.Context, status uint8) (int64, error)
}

type reportRepository struct {
	db *database.DB
}

func NewReportRepository(db *database.DB) ReportRepository {
	return &reportRepository{db: db}
}

func (r *reportRepository) Create(ctx context.Context, item *model.Report) error {
	return database.Conn(ctx, r.db).Create(item).Error
}

func (r *reportRepository) Update(ctx context.Context, item *model.Report) error {
	return database.Conn(ctx, r.db).Save(item).Error
}

func (r *reportRepository) GetByID(ctx context.Context, id uint) (*model.Report, error) {
	var item model.Report
	if err := database.Conn(ctx, r.db).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func applyReportListFilter(db *gorm.DB, f ReportListFilter) *gorm.DB {
	if f.ReporterID != nil {
		db = db.Where("reporter_id = ?", *f.ReporterID)
	}
	if f.Status != nil {
		db = db.Where("`status` = ?", *f.Status)
	}
	if f.TargetType != nil {
		db = db.Where("target_type = ?", *f.TargetType)
	}
	if f.Reason != nil {
		db = db.Where("reason = ?", *f.Reason)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		db = db.Where("content LIKE ? OR handle_result LIKE ?", kw, kw)
	}
	return db
}

func (r *reportRepository) List(ctx context.Context, f ReportListFilter) ([]*model.Report, int64, error) {
	db := applyReportListFilter(database.Conn(ctx, r.db).Model(&model.Report{}), f)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.Report
	if err := db.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *reportRepository) CountByReporterTargetSince(ctx context.Context, reporterID uint, targetType uint8, targetID uint, since time.Time) (int64, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.Report{}).
		Where("reporter_id = ? AND target_type = ? AND target_id = ? AND created_at >= ?", reporterID, targetType, targetID, since).
		Count(&count).Error
	return count, err
}

func (r *reportRepository) CountByReporterBetween(ctx context.Context, reporterID uint, start, end time.Time) (int64, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.Report{}).
		Where("reporter_id = ? AND created_at >= ? AND created_at < ?", reporterID, start, end).
		Count(&count).Error
	return count, err
}

func (r *reportRepository) CountByTarget(ctx context.Context, targetType uint8, targetID uint) (int64, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.Report{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&count).Error
	return count, err
}

func (r *reportRepository) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Report{}).Where("status = ?", status).Count(&n).Error
	return n, err
}
