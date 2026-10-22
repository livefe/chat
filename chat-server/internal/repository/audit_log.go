package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type AuditLogListFilter struct {
	EmployeeID  *uint
	Module      string
	Action      string
	TargetType  string
	TargetID    *uint
	RequestPath string
	Success     *bool
	Keyword     string
	StartTime   *time.Time
	EndTime     *time.Time
	Page        int
	PageSize    int
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetByID(ctx context.Context, id uint) (*model.AuditLog, error)
	List(ctx context.Context, f AuditLogListFilter) ([]*model.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *database.DB
}

func NewAuditLogRepository(db *database.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	return database.Conn(ctx, r.db).Create(log).Error
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uint) (*model.AuditLog, error) {
	var row model.AuditLog
	if err := database.Conn(ctx, r.db).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func applyAuditLogListFilter(db *gorm.DB, f AuditLogListFilter) *gorm.DB {
	if f.EmployeeID != nil {
		db = db.Where("employee_id = ?", *f.EmployeeID)
	}
	if f.Module != "" {
		db = db.Where("module = ?", f.Module)
	}
	if f.Action != "" {
		db = db.Where("action = ?", f.Action)
	}
	if f.TargetType != "" {
		db = db.Where("target_type = ?", f.TargetType)
	}
	if f.TargetID != nil {
		db = db.Where("target_id = ?", *f.TargetID)
	}
	if f.RequestPath != "" {
		db = db.Where("request_path LIKE ?", "%"+f.RequestPath+"%")
	}
	if f.Success != nil {
		db = db.Where("success = ?", *f.Success)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		db = db.Where(
			"employee_name LIKE ? OR employee_mobile LIKE ? OR target_label LIKE ? OR action_text LIKE ? OR request_path LIKE ? OR detail LIKE ?",
			kw, kw, kw, kw, kw, kw,
		)
	}
	if f.StartTime != nil {
		db = db.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		db = db.Where("created_at <= ?", *f.EndTime)
	}
	return db
}

func (r *auditLogRepository) List(ctx context.Context, f AuditLogListFilter) ([]*model.AuditLog, int64, error) {
	db := applyAuditLogListFilter(database.Conn(ctx, r.db).Model(&model.AuditLog{}), f)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.AuditLog
	if err := db.Order("created_at DESC, id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
