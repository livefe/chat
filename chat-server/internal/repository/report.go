package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type ReportListFilter struct {
	ReporterID   uint
	TargetUserID uint
	TargetID     uint
	Status       *uint8
	TargetType   *uint8
	Reason       *uint8
	Keyword      string
	StartTime    string
	EndTime      string
	OldestFirst  bool
	Page         int
	PageSize     int
}

type ReportTargetKey struct {
	TargetType uint8
	TargetID   uint
}

type ReportRepository interface {
	Create(ctx context.Context, r *model.Report) error
	Update(ctx context.Context, r *model.Report) error
	GetByID(ctx context.Context, id uint) (*model.Report, error)
	List(ctx context.Context, f ReportListFilter) ([]*model.Report, int64, error)
	CountByReporterTargetSince(ctx context.Context, reporterID uint, targetType uint8, targetID uint, since time.Time) (int64, error)
	CountByReporterBetween(ctx context.Context, reporterID uint, start, end time.Time) (int64, error)
	CountByTargets(ctx context.Context, keys []ReportTargetKey) (map[ReportTargetKey]int64, error)
	CountByTargetUser(ctx context.Context, userID uint) (int64, error)
	CountByStatus(ctx context.Context, status uint8) (int64, error)
	ResolvePendingByTarget(ctx context.Context, key ReportTargetKey, excludeID uint, updates map[string]any) (int64, error)
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
	if f.ReporterID > 0 {
		db = db.Where("reporter_id = ?", f.ReporterID)
	}
	if f.TargetUserID > 0 {
		db = db.Where("target_user_id = ?", f.TargetUserID)
	}
	if f.TargetID > 0 {
		db = db.Where("target_id = ?", f.TargetID)
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
	if f.StartTime != "" {
		db = db.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime != "" {
		db = db.Where("created_at <= ?", f.EndTime)
	}
	return db
}

func (r *reportRepository) List(ctx context.Context, f ReportListFilter) ([]*model.Report, int64, error) {
	db := applyReportListFilter(database.Conn(ctx, r.db).Model(&model.Report{}), f)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "id DESC"
	if f.OldestFirst {
		order = "id ASC"
	}
	var list []*model.Report
	if err := db.Order(order).
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

func (r *reportRepository) CountByTargets(ctx context.Context, keys []ReportTargetKey) (map[ReportTargetKey]int64, error) {
	out := make(map[ReportTargetKey]int64, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	pairs := make([][]any, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, []any{k.TargetType, k.TargetID})
	}
	var rows []struct {
		TargetType uint8
		TargetID   uint
		Total      int64
	}
	err := database.Conn(ctx, r.db).Model(&model.Report{}).
		Select("target_type, target_id, COUNT(*) AS total").
		Where("(target_type, target_id) IN ?", pairs).
		Group("target_type, target_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[ReportTargetKey{TargetType: row.TargetType, TargetID: row.TargetID}] = row.Total
	}
	return out, nil
}

func (r *reportRepository) CountByTargetUser(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Report{}).Where("target_user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *reportRepository) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Report{}).Where("status = ?", status).Count(&n).Error
	return n, err
}

func (r *reportRepository) ResolvePendingByTarget(ctx context.Context, key ReportTargetKey, excludeID uint, updates map[string]any) (int64, error) {
	res := database.Conn(ctx, r.db).Model(&model.Report{}).
		Where("target_type = ? AND target_id = ? AND id <> ? AND `status` = ?",
			key.TargetType, key.TargetID, excludeID, model.ReportStatusPending).
		Updates(updates)
	return res.RowsAffected, res.Error
}
