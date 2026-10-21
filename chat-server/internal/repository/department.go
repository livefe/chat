package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type DepartmentListFilter struct {
	Name   string
	Status *uint8
}

type DepartmentRepository interface {
	Create(ctx context.Context, dept *model.Department) error
	Update(ctx context.Context, dept *model.Department) error
	GetByID(ctx context.Context, id uint) (*model.Department, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.Department, error)
	List(ctx context.Context, f DepartmentListFilter) ([]*model.Department, error)
	ListEnabled(ctx context.Context) ([]*model.Department, error)
	ListByPathPrefix(ctx context.Context, pathPrefix string) ([]*model.Department, error)
	CountChildren(ctx context.Context, parentID uint) (int64, error)
}

type departmentRepository struct {
	db *database.DB
}

func NewDepartmentRepository(db *database.DB) DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, dept *model.Department) error {
	return database.Conn(ctx, r.db).Create(dept).Error
}

func (r *departmentRepository) Update(ctx context.Context, dept *model.Department) error {
	return database.Conn(ctx, r.db).Save(dept).Error
}

func (r *departmentRepository) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	var dept model.Department
	if err := database.Conn(ctx, r.db).First(&dept, id).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *departmentRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.Department, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*model.Department
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *departmentRepository) List(ctx context.Context, f DepartmentListFilter) ([]*model.Department, error) {
	db := database.Conn(ctx, r.db).Model(&model.Department{})
	if f.Name != "" {
		db = db.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	var list []*model.Department
	if err := db.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *departmentRepository) ListEnabled(ctx context.Context) ([]*model.Department, error) {
	var list []*model.Department
	err := database.Conn(ctx, r.db).
		Where("status = ?", model.DepartmentStatusEnabled).
		Order("sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *departmentRepository) ListByPathPrefix(ctx context.Context, pathPrefix string) ([]*model.Department, error) {
	if pathPrefix == "" {
		return nil, nil
	}
	var list []*model.Department
	err := database.Conn(ctx, r.db).
		Where("path LIKE ?", pathPrefix+"%").
		Order("level ASC, sort ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *departmentRepository) CountChildren(ctx context.Context, parentID uint) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Department{}).Where("parent_id = ?", parentID).Count(&n).Error
	return n, err
}
