package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type EmployeeRoleRepository interface {
	DeleteByRoleID(ctx context.Context, roleID uint) error
	ListRoleIDsByEmployeeID(ctx context.Context, employeeID uint) ([]uint, error)
	ListRoleIDsByEmployeeIDs(ctx context.Context, employeeIDs []uint) (map[uint][]uint, error)
	HasAnyRole(ctx context.Context, employeeID uint) (bool, error)
	ReplaceByEmployeeID(ctx context.Context, employeeID uint, items []*model.EmployeeRole) error
}

type employeeRoleRepository struct {
	db *database.DB
}

func NewEmployeeRoleRepository(db *database.DB) EmployeeRoleRepository {
	return &employeeRoleRepository{db: db}
}

func (r *employeeRoleRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return database.Conn(ctx, r.db).Where("role_id = ?", roleID).Delete(&model.EmployeeRole{}).Error
}

func (r *employeeRoleRepository) ListRoleIDsByEmployeeID(ctx context.Context, employeeID uint) ([]uint, error) {
	var ids []uint
	err := database.Conn(ctx, r.db).Model(&model.EmployeeRole{}).
		Where("employee_id = ?", employeeID).
		Pluck("role_id", &ids).Error
	return ids, err
}

func (r *employeeRoleRepository) ListRoleIDsByEmployeeIDs(ctx context.Context, employeeIDs []uint) (map[uint][]uint, error) {
	out := make(map[uint][]uint)
	if len(employeeIDs) == 0 {
		return out, nil
	}

	type row struct {
		EmployeeID uint
		RoleID     uint
	}
	var rows []row
	if err := database.Conn(ctx, r.db).Model(&model.EmployeeRole{}).
		Select("employee_id, role_id").
		Where("employee_id IN ?", employeeIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, item := range rows {
		out[item.EmployeeID] = append(out[item.EmployeeID], item.RoleID)
	}
	return out, nil
}

func (r *employeeRoleRepository) HasAnyRole(ctx context.Context, employeeID uint) (bool, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.EmployeeRole{}).
		Where("employee_id = ?", employeeID).
		Count(&count).Error
	return count > 0, err
}

func (r *employeeRoleRepository) ReplaceByEmployeeID(ctx context.Context, employeeID uint, items []*model.EmployeeRole) error {
	return database.Conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("employee_id = ?", employeeID).Delete(&model.EmployeeRole{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(items).Error
	})
}
