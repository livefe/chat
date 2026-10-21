package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type RoleListFilter struct {
	Name     string
	Code     string
	Status   *uint8
	Page     int
	PageSize int
}

type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.Role, error)
	GetByCode(ctx context.Context, code string) (*model.Role, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.Role, error)
	List(ctx context.Context, f RoleListFilter) ([]*model.Role, int64, error)
}

type roleRepository struct {
	db *database.DB
}

func NewRoleRepository(db *database.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	return database.Conn(ctx, r.db).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	return database.Conn(ctx, r.db).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Delete(&model.Role{}, id).Error
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	if err := database.Conn(ctx, r.db).First(&role, id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByCode(ctx context.Context, code string) (*model.Role, error) {
	var role model.Role
	if err := database.Conn(ctx, r.db).Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var roles []*model.Role
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) List(ctx context.Context, f RoleListFilter) ([]*model.Role, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.Role{})

	if f.Name != "" {
		db = db.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Code != "" {
		db = db.Where("code LIKE ?", "%"+f.Code+"%")
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []*model.Role
	if err := db.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}
