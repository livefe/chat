package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type EmployeeLoginRecordRepository interface {
	Create(ctx context.Context, record *model.EmployeeLoginRecord) error
	ListByEmployee(ctx context.Context, employeeID uint, mobile string, page, pageSize int) ([]*model.EmployeeLoginRecord, int64, error)
	LatestSuccessByEmployeeIDs(ctx context.Context, employeeIDs []uint) (map[uint]*model.EmployeeLoginRecord, error)
}

type employeeLoginRecordRepository struct {
	db *database.DB
}

func NewEmployeeLoginRecordRepository(db *database.DB) EmployeeLoginRecordRepository {
	return &employeeLoginRecordRepository{db: db}
}

func (r *employeeLoginRecordRepository) Create(ctx context.Context, record *model.EmployeeLoginRecord) error {
	return database.Conn(ctx, r.db).Create(record).Error
}

func (r *employeeLoginRecordRepository) ListByEmployee(ctx context.Context, employeeID uint, mobile string, page, pageSize int) ([]*model.EmployeeLoginRecord, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.EmployeeLoginRecord{}).
		Where("employee_id = ? OR mobile = ?", employeeID, mobile)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.EmployeeLoginRecord
	err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *employeeLoginRecordRepository) LatestSuccessByEmployeeIDs(ctx context.Context, employeeIDs []uint) (map[uint]*model.EmployeeLoginRecord, error) {
	out := make(map[uint]*model.EmployeeLoginRecord, len(employeeIDs))
	if len(employeeIDs) == 0 {
		return out, nil
	}
	conn := database.Conn(ctx, r.db)
	latest := conn.Model(&model.EmployeeLoginRecord{}).
		Select("MAX(id)").
		Where("employee_id IN ? AND status = ?", employeeIDs, model.SMSStatusSuccess).
		Group("employee_id")
	var list []*model.EmployeeLoginRecord
	if err := conn.Where("id IN (?)", latest).Find(&list).Error; err != nil {
		return nil, err
	}
	for _, row := range list {
		out[row.EmployeeID] = row
	}
	return out, nil
}
