package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type RolePermissionRepository interface {
	DeleteByRoleID(ctx context.Context, roleID uint) error
	GetPermissionIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error)
	GetPermissionIDsByRoleIDs(ctx context.Context, roleIDs []uint) ([]uint, error)
	CountByPermissionID(ctx context.Context, permissionID uint) (int64, error)
	ReplaceByRoleID(ctx context.Context, roleID uint, items []*model.RolePermission) error
}

type rolePermissionRepository struct {
	db *database.DB
}

func NewRolePermissionRepository(db *database.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	return database.Conn(ctx, r.db).Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error
}

func (r *rolePermissionRepository) GetPermissionIDsByRoleID(ctx context.Context, roleID uint) ([]uint, error) {
	var ids []uint
	err := database.Conn(ctx, r.db).Model(&model.RolePermission{}).
		Where("role_id = ?", roleID).
		Pluck("permission_id", &ids).Error
	return ids, err
}

func (r *rolePermissionRepository) GetPermissionIDsByRoleIDs(ctx context.Context, roleIDs []uint) ([]uint, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var ids []uint
	err := database.Conn(ctx, r.db).Model(&model.RolePermission{}).
		Where("role_id IN ?", roleIDs).
		Distinct("permission_id").
		Pluck("permission_id", &ids).Error
	return ids, err
}

func (r *rolePermissionRepository) CountByPermissionID(ctx context.Context, permissionID uint) (int64, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.RolePermission{}).
		Where("permission_id = ?", permissionID).
		Count(&count).Error
	return count, err
}

func (r *rolePermissionRepository) ReplaceByRoleID(ctx context.Context, roleID uint, items []*model.RolePermission) error {
	return database.Conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(items).Error
	})
}
