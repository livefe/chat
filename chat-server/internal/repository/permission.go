package repository

import (
	"context"
	"strings"

	"app/internal/model"
	"app/pkg/database"
)

type PermissionRepository interface {
	Create(ctx context.Context, p *model.Permission) error
	Update(ctx context.Context, p *model.Permission) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.Permission, error)
	GetAssignableByCode(ctx context.Context, code string) (*model.Permission, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.Permission, error)
	ListByParentID(ctx context.Context, parentID uint) ([]*model.Permission, error)
	ListAll(ctx context.Context) ([]*model.Permission, error)
	GetAPIByPathAndMethod(ctx context.Context, path, method string) (*model.Permission, error)
	ListEnabledAPIs(ctx context.Context) ([]*model.Permission, error)
}

type permissionRepository struct {
	db *database.DB
}

func NewPermissionRepository(db *database.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) Create(ctx context.Context, p *model.Permission) error {
	return database.Conn(ctx, r.db).Create(p).Error
}

func (r *permissionRepository) Update(ctx context.Context, p *model.Permission) error {
	return database.Conn(ctx, r.db).Save(p).Error
}

func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Delete(&model.Permission{}, id).Error
}

func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*model.Permission, error) {
	var p model.Permission
	if err := database.Conn(ctx, r.db).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *permissionRepository) GetAssignableByCode(ctx context.Context, code string) (*model.Permission, error) {
	var p model.Permission
	if err := database.Conn(ctx, r.db).
		Where("code = ? AND type IN ?", code, []uint8{model.PermissionTypeMenu, model.PermissionTypeButton}).
		First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *permissionRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.Permission, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var ps []*model.Permission
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *permissionRepository) ListByParentID(ctx context.Context, parentID uint) ([]*model.Permission, error) {
	var ps []*model.Permission
	if err := database.Conn(ctx, r.db).Where("parent_id = ?", parentID).Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *permissionRepository) ListAll(ctx context.Context) ([]*model.Permission, error) {
	var ps []*model.Permission
	if err := database.Conn(ctx, r.db).Order("sort ASC, created_at DESC").Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}

func (r *permissionRepository) GetAPIByPathAndMethod(ctx context.Context, path, method string) (*model.Permission, error) {
	reqPath := strings.TrimSuffix(strings.TrimSpace(path), "/")
	reqMethod := strings.ToUpper(strings.TrimSpace(method))

	var p model.Permission
	if err := database.Conn(ctx, r.db).
		Where(
			"type = ? AND status = ? AND path = ? AND UPPER(method) = ?",
			model.PermissionTypeAPI,
			model.PermissionStatusEnabled,
			reqPath,
			reqMethod,
		).
		First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *permissionRepository) ListEnabledAPIs(ctx context.Context) ([]*model.Permission, error) {
	var ps []*model.Permission
	if err := database.Conn(ctx, r.db).
		Where("type = ? AND status = ?", model.PermissionTypeAPI, model.PermissionStatusEnabled).
		Order("sort ASC, id ASC").
		Find(&ps).Error; err != nil {
		return nil, err
	}
	return ps, nil
}
