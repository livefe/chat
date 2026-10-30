package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var employeeStatusText = map[uint8]string{
	model.EmployeeStatusDisabled: "禁用",
	model.EmployeeStatusEnabled:  "启用",
}

type EmployeeService struct {
	log              *logger.Logger
	authSvc          *AuthService
	permSvc          *PermissionService
	mediaSvc         *MediaService
	employeeRepo     repository.EmployeeRepository
	employeeRoleRepo repository.EmployeeRoleRepository
	loginRecordRepo  repository.EmployeeLoginRecordRepository
	roleRepo         repository.RoleRepository
	deptRepo         repository.DepartmentRepository
}

func NewEmployeeService(
	log *logger.Logger,
	authSvc *AuthService,
	permSvc *PermissionService,
	mediaSvc *MediaService,
	employeeRepo repository.EmployeeRepository,
	employeeRoleRepo repository.EmployeeRoleRepository,
	loginRecordRepo repository.EmployeeLoginRecordRepository,
	roleRepo repository.RoleRepository,
	deptRepo repository.DepartmentRepository,
) *EmployeeService {
	return &EmployeeService{
		log:              log,
		authSvc:          authSvc,
		permSvc:          permSvc,
		mediaSvc:         mediaSvc,
		employeeRepo:     employeeRepo,
		employeeRoleRepo: employeeRoleRepo,
		loginRecordRepo:  loginRecordRepo,
		roleRepo:         roleRepo,
		deptRepo:         deptRepo,
	}
}

func employeeAvatarURLs(ctx context.Context, mediaSvc *MediaService, employees []*model.Employee) map[uint]string {
	out := make(map[uint]string, len(employees))
	if mediaSvc == nil {
		return out
	}
	mediaIDs := make([]uint, 0, len(employees))
	for _, e := range employees {
		if e != nil && e.AvatarMediaID > 0 {
			mediaIDs = append(mediaIDs, e.AvatarMediaID)
		}
	}
	urls := mediaSvc.ResolveURLs(ctx, mediaIDs)
	for _, e := range employees {
		if e == nil {
			continue
		}
		if u := urls[e.AvatarMediaID]; u != "" {
			out[e.ID] = u
		}
	}
	return out
}

func toEmployeeItem(e *model.Employee, roles []types.EmployeeRoleBrief, departmentName, avatarURL string, lastLogin *model.EmployeeLoginRecord) types.EmployeeItem {
	item := types.EmployeeItem{
		ID:           e.ID,
		Mobile:       e.Mobile,
		Name:         e.Name,
		EmployeeNo:   e.EmployeeNo,
		AvatarURL:    avatarURL,
		DepartmentID: e.DepartmentID,
		Department:   departmentName,
		Roles:        roles,
		Status:       e.Status,
		StatusText:   utils.TextOf(employeeStatusText, e.Status),
		Remark:       e.Remark,
		CreatedAt:    utils.FormatDateTime(e.CreatedAt),
		UpdatedAt:    utils.FormatDateTime(e.UpdatedAt),
	}
	if lastLogin != nil {
		item.LastLoginAt = utils.FormatDateTime(lastLogin.CreatedAt)
		item.LastLoginIP = lastLogin.IP
	}
	return item
}

func (s *EmployeeService) resolveDepartment(ctx context.Context, departmentID uint) (uint, string, error) {
	if departmentID == 0 {
		return 0, "", errs.Invalid("请选择部门")
	}
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", errs.NotFound("部门不存在")
		}
		return 0, "", fmt.Errorf("查询部门失败: %w", err)
	}
	if dept.Status != model.DepartmentStatusEnabled {
		return 0, "", errs.Invalid("部门已禁用")
	}
	return dept.ID, dept.Name, nil
}

func toEmployeeInfo(e *model.Employee, departmentName, avatarURL string) types.EmployeeInfo {
	return types.EmployeeInfo{
		ID:            e.ID,
		Mobile:        utils.DesensitizeMobile(e.Mobile),
		Name:          e.Name,
		EmployeeNo:    e.EmployeeNo,
		AvatarMediaID: e.AvatarMediaID,
		AvatarURL:     avatarURL,
		Department:    departmentName,
		Status:        e.Status,
		Remark:        e.Remark,
	}
}

func (s *EmployeeService) employeeInfo(ctx context.Context, e *model.Employee) types.EmployeeInfo {
	departmentName := ""
	if dept, err := s.deptRepo.GetByID(ctx, e.DepartmentID); err == nil {
		departmentName = dept.Name
	}
	return toEmployeeInfo(e, departmentName, s.mediaSvc.ResolveURL(ctx, e.AvatarMediaID))
}

func (s *EmployeeService) EnsureActive(ctx context.Context, employeeID uint) error {
	if employeeID == 0 {
		return errs.Unauthorized("员工ID无效")
	}
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.Unauthorized("员工不存在")
		}
		return errs.Unauthorized("认证失败")
	}
	if employee.Status != model.EmployeeStatusEnabled {
		return errs.Unauthorized("账号已被禁用")
	}
	return nil
}

func (s *EmployeeService) Get(ctx context.Context, employeeID uint) (*types.EmployeeInfo, error) {
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("员工不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	info := s.employeeInfo(ctx, employee)
	return &info, nil
}

func (s *EmployeeService) UpdateProfile(ctx context.Context, employeeID uint, req *types.UpdateMyProfileRequest) (*types.EmployeeInfo, error) {
	employee, err := s.employeeRepo.GetByID(ctx, employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("员工不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if employee.Status != model.EmployeeStatusEnabled {
		return nil, errs.Forbidden("账号已被禁用")
	}

	oldID, newID := employee.AvatarMediaID, req.AvatarMediaID
	employee.AvatarMediaID = newID
	persist := func() error {
		if err := s.employeeRepo.Update(ctx, employee); err != nil {
			s.log.Error(ctx, "更新个人资料失败", logger.Error(err))
			return fmt.Errorf("保存失败: %w", err)
		}
		return nil
	}
	if err := s.mediaSvc.ReplaceSingle(ctx, EmployeeUploader(employeeID), oldID, newID, model.MediaKindImage, model.MediaBizEmployeeAvatar, employeeID, persist); err != nil {
		return nil, err
	}
	info := s.employeeInfo(ctx, employee)
	return &info, nil
}
