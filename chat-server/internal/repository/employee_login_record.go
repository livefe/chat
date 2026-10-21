package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type EmployeeLoginRecordRepository interface {
	Create(ctx context.Context, record *model.EmployeeLoginRecord) error
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
