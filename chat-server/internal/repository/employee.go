package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type EmployeeListFilter struct {
	ID            *uint
	Keyword       string
	Mobile        string
	Name          string
	EmployeeNo    string
	DepartmentID  *uint
	DepartmentIDs []uint
	RoleID        *uint
	Status        *uint8
	ExcludeID     uint
	StartTime     string
	EndTime       string
	Page          int
	PageSize      int
}

type EmployeeRepository interface {
	Create(ctx context.Context, employee *model.Employee) error
	Update(ctx context.Context, employee *model.Employee) error
	GetByID(ctx context.Context, id uint) (*model.Employee, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.Employee, error)
	GetByMobile(ctx context.Context, mobile string) (*model.Employee, error)
	List(ctx context.Context, f EmployeeListFilter) ([]*model.Employee, int64, error)
	CountByDepartmentID(ctx context.Context, departmentID uint) (int64, error)
	CountGroupByDepartmentIDs(ctx context.Context, departmentIDs []uint) (map[uint]int64, error)
}

type employeeRepository struct {
	db *database.DB
}

func NewEmployeeRepository(db *database.DB) EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, employee *model.Employee) error {
	return database.Conn(ctx, r.db).Create(employee).Error
}

func (r *employeeRepository) Update(ctx context.Context, employee *model.Employee) error {
	return database.Conn(ctx, r.db).Save(employee).Error
}

func (r *employeeRepository) GetByID(ctx context.Context, id uint) (*model.Employee, error) {
	var employee model.Employee
	if err := database.Conn(ctx, r.db).First(&employee, id).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.Employee, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var employees []*model.Employee
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

func (r *employeeRepository) GetByMobile(ctx context.Context, mobile string) (*model.Employee, error) {
	var employee model.Employee
	if err := database.Conn(ctx, r.db).Where("mobile = ?", mobile).First(&employee).Error; err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *employeeRepository) List(ctx context.Context, f EmployeeListFilter) ([]*model.Employee, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.Employee{})
	if f.ID != nil && *f.ID > 0 {
		db = db.Where("id = ?", *f.ID)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		db = db.Where("name LIKE ? OR mobile LIKE ? OR employee_no LIKE ?", kw, kw, kw)
	}
	if f.Mobile != "" {
		db = db.Where("mobile LIKE ?", "%"+f.Mobile+"%")
	}
	if f.Name != "" {
		db = db.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.EmployeeNo != "" {
		db = db.Where("employee_no LIKE ?", "%"+f.EmployeeNo+"%")
	}
	if f.RoleID != nil && *f.RoleID > 0 {
		db = db.Where("id IN (?)", database.Conn(ctx, r.db).Model(&model.EmployeeRole{}).
			Select("employee_id").Where("role_id = ?", *f.RoleID))
	}
	if len(f.DepartmentIDs) > 0 {
		db = db.Where("department_id IN ?", f.DepartmentIDs)
	} else if f.DepartmentID != nil && *f.DepartmentID > 0 {
		db = db.Where("department_id = ?", *f.DepartmentID)
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.ExcludeID > 0 {
		db = db.Where("id <> ?", f.ExcludeID)
	}
	if f.StartTime != "" {
		db = db.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime != "" {
		db = db.Where("created_at <= ?", f.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var employees []*model.Employee
	offset := (f.Page - 1) * f.PageSize
	if err := db.Order("id DESC").Offset(offset).Limit(f.PageSize).Find(&employees).Error; err != nil {
		return nil, 0, err
	}
	return employees, total, nil
}

func (r *employeeRepository) CountByDepartmentID(ctx context.Context, departmentID uint) (int64, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.Employee{}).
		Where("department_id = ?", departmentID).
		Count(&count).Error
	return count, err
}

func (r *employeeRepository) CountGroupByDepartmentIDs(ctx context.Context, departmentIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64)
	if len(departmentIDs) == 0 {
		return result, nil
	}
	type row struct {
		DepartmentID uint  `gorm:"column:department_id"`
		Count        int64 `gorm:"column:count"`
	}
	var rows []row
	err := database.Conn(ctx, r.db).Model(&model.Employee{}).
		Select("department_id, COUNT(*) AS count").
		Where("department_id IN ?", departmentIDs).
		Group("department_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.DepartmentID] = item.Count
	}
	return result, nil
}
