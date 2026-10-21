package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/auth"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *EmployeeService) AdminList(ctx context.Context, req *types.EmployeeListRequest) (*types.EmployeeListResponse, error) {
	filter := repository.EmployeeListFilter{
		Mobile:    req.Mobile,
		Name:      req.Name,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Page:      req.Page,
		PageSize:  req.PageSize,
	}
	if req.DepartmentID != nil && *req.DepartmentID > 0 {
		ids, err := s.departmentSubtreeIDs(ctx, *req.DepartmentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &types.EmployeeListResponse{Total: 0, List: []types.EmployeeItem{}}, nil
			}
			s.log.Error(ctx, "查询部门子树失败", logger.Error(err))
			return nil, fmt.Errorf("查询员工列表失败: %w", err)
		}
		filter.DepartmentIDs = ids
	}
	employees, total, err := s.employeeRepo.List(ctx, filter)
	if err != nil {
		s.log.Error(ctx, "查询员工列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询员工列表失败: %w", err)
	}

	departmentNameByID, err := s.employeeDepartmentNames(ctx, employees)
	if err != nil {
		s.log.Error(ctx, "查询部门详情失败", logger.Error(err))
		return nil, fmt.Errorf("查询员工列表失败: %w", err)
	}
	rolesByEmployee, err := s.employeeRoleBriefs(ctx, employees)
	if err != nil {
		s.log.Error(ctx, "查询员工角色失败", logger.Error(err))
		return nil, fmt.Errorf("查询员工列表失败: %w", err)
	}

	list := make([]types.EmployeeItem, 0, len(employees))
	for _, e := range employees {
		list = append(list, toEmployeeItem(e, rolesByEmployee[e.ID], departmentNameByID[e.DepartmentID]))
	}
	return &types.EmployeeListResponse{Total: total, List: list}, nil
}

func (s *EmployeeService) employeeDepartmentNames(ctx context.Context, employees []*model.Employee) (map[uint]string, error) {
	departmentIDs := make([]uint, 0, len(employees))
	seen := make(map[uint]struct{})
	for _, e := range employees {
		if e.DepartmentID == 0 {
			continue
		}
		if _, ok := seen[e.DepartmentID]; !ok {
			seen[e.DepartmentID] = struct{}{}
			departmentIDs = append(departmentIDs, e.DepartmentID)
		}
	}
	names := make(map[uint]string, len(departmentIDs))
	if len(departmentIDs) == 0 {
		return names, nil
	}
	depts, err := s.deptRepo.GetByIDs(ctx, departmentIDs)
	if err != nil {
		return nil, err
	}
	for _, d := range depts {
		names[d.ID] = d.Name
	}
	return names, nil
}

func (s *EmployeeService) employeeRoleBriefs(ctx context.Context, employees []*model.Employee) (map[uint][]types.EmployeeRoleBrief, error) {
	employeeIDs := make([]uint, 0, len(employees))
	for _, e := range employees {
		employeeIDs = append(employeeIDs, e.ID)
	}
	roleIDsByEmployee, err := s.employeeRoleRepo.ListRoleIDsByEmployeeIDs(ctx, employeeIDs)
	if err != nil {
		return nil, fmt.Errorf("查询员工角色关联失败: %w", err)
	}

	roleIDSet := make(map[uint]struct{})
	for _, roleIDs := range roleIDsByEmployee {
		for _, rid := range roleIDs {
			roleIDSet[rid] = struct{}{}
		}
	}
	roleIDs := make([]uint, 0, len(roleIDSet))
	for rid := range roleIDSet {
		roleIDs = append(roleIDs, rid)
	}
	roleByID := make(map[uint]*model.Role, len(roleIDs))
	if len(roleIDs) > 0 {
		roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
		if err != nil {
			return nil, fmt.Errorf("查询角色详情失败: %w", err)
		}
		for _, r := range roles {
			roleByID[r.ID] = r
		}
	}

	out := make(map[uint][]types.EmployeeRoleBrief, len(employees))
	for _, e := range employees {
		roles := make([]types.EmployeeRoleBrief, 0)
		for _, rid := range roleIDsByEmployee[e.ID] {
			if r, ok := roleByID[rid]; ok {
				roles = append(roles, types.EmployeeRoleBrief{ID: r.ID, Name: r.Name, Code: r.Code})
			}
		}
		sort.SliceStable(roles, func(i, j int) bool {
			return roles[i].ID < roles[j].ID
		})
		out[e.ID] = roles
	}
	return out, nil
}

func (s *EmployeeService) AdminCreate(ctx context.Context, req *types.CreateEmployeeRequest) (uint, error) {
	existing, err := s.employeeRepo.GetByMobile(ctx, req.Mobile)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if existing != nil {
		return 0, errs.Conflict("手机号已存在")
	}

	deptID, _, err := s.resolveDepartment(ctx, req.DepartmentID)
	if err != nil {
		return 0, err
	}

	employee := &model.Employee{
		Mobile:       req.Mobile,
		Name:         req.Name,
		EmployeeNo:   req.EmployeeNo,
		AvatarURL:    req.AvatarURL,
		DepartmentID: deptID,
		Status:       req.Status,
		Remark:       req.Remark,
	}
	if err := s.employeeRepo.Create(ctx, employee); err != nil {
		s.log.Error(ctx, "创建员工失败", logger.Error(err))
		return 0, fmt.Errorf("创建员工失败: %w", err)
	}
	return employee.ID, nil
}

func (s *EmployeeService) AdminUpdate(ctx context.Context, req *types.UpdateEmployeeRequest) error {
	employee, err := s.employeeRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("员工不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}

	if req.Mobile != employee.Mobile {
		other, err := s.employeeRepo.GetByMobile(ctx, req.Mobile)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		if other != nil && other.ID != employee.ID {
			return errs.Conflict("手机号已被其他员工使用")
		}
	}

	employee.Mobile = req.Mobile
	employee.Name = req.Name
	employee.EmployeeNo = req.EmployeeNo
	employee.AvatarURL = req.AvatarURL
	deptID, _, err := s.resolveDepartment(ctx, req.DepartmentID)
	if err != nil {
		return err
	}
	employee.DepartmentID = deptID
	employee.Status = req.Status
	employee.Remark = req.Remark
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		s.log.Error(ctx, "更新员工失败", logger.Error(err))
		return fmt.Errorf("更新员工失败: %w", err)
	}
	return nil
}

func (s *EmployeeService) AdminUpdateStatus(ctx context.Context, operatorID, id uint, status uint8) error {
	if operatorID == id && status != model.EmployeeStatusEnabled {
		return errs.Forbidden("不能禁用自己的账号")
	}
	employee, err := s.employeeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("员工不存在")
		}
		return fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if status != model.EmployeeStatusEnabled {
		isTargetSuper, err := s.permSvc.IsSuperAdmin(ctx, id)
		if err != nil {
			return fmt.Errorf("校验权限失败: %w", err)
		}
		if isTargetSuper {
			return errs.Forbidden("不能禁用超级管理员")
		}
	}
	employee.Status = status
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		s.log.Error(ctx, "更新员工状态失败", logger.Error(err))
		return fmt.Errorf("更新员工状态失败: %w", err)
	}
	if status != model.EmployeeStatusEnabled {
		s.authSvc.Revoke(ctx, auth.SubjectTypeEmployee, id)
	}
	return nil
}

func (s *EmployeeService) AdminListRoles(ctx context.Context, employeeID uint) ([]types.RoleItem, error) {
	roleIDs, err := s.employeeRoleRepo.ListRoleIDsByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("获取员工角色失败: %w", err)
	}
	if len(roleIDs) == 0 {
		return []types.RoleItem{}, nil
	}

	roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
	if err != nil {
		return nil, fmt.Errorf("获取角色详情失败: %w", err)
	}

	items := make([]types.RoleItem, 0, len(roles))
	for _, r := range roles {
		items = append(items, toRoleItem(r))
	}
	return items, nil
}

func (s *EmployeeService) AdminAssignRoles(ctx context.Context, operatorID uint, req *types.AssignEmployeeRolesRequest) error {
	if _, err := s.employeeRepo.GetByID(ctx, req.EmployeeID); err != nil {
		return errs.NotFound("员工不存在")
	}

	existingRoleIDs, err := s.employeeRoleRepo.ListRoleIDsByEmployeeID(ctx, req.EmployeeID)
	if err != nil {
		return fmt.Errorf("查询员工角色失败: %w", err)
	}
	targetHadSuperAdmin := false
	if len(existingRoleIDs) > 0 {
		existingRoles, err := s.roleRepo.GetByIDs(ctx, existingRoleIDs)
		if err != nil {
			return fmt.Errorf("查询角色失败: %w", err)
		}
		for _, role := range existingRoles {
			if role.Code == types.SuperAdminRoleCode {
				targetHadSuperAdmin = true
				break
			}
		}
	}

	containsSuperAdminRole := false
	if len(req.RoleIDs) > 0 {
		roles, err := s.roleRepo.GetByIDs(ctx, req.RoleIDs)
		if err != nil {
			return fmt.Errorf("查询角色失败: %w", err)
		}
		if len(roles) != len(req.RoleIDs) {
			return errs.NotFound("部分角色不存在")
		}
		for _, role := range roles {
			if role.Status != model.RoleStatusEnabled {
				return errs.Invalid("不能分配已禁用角色")
			}
			if role.Code == types.SuperAdminRoleCode {
				containsSuperAdminRole = true
			}
		}
	}

	if containsSuperAdminRole || targetHadSuperAdmin {
		isOperatorSuperAdmin, err := s.permSvc.IsSuperAdmin(ctx, operatorID)
		if err != nil {
			return fmt.Errorf("校验操作者权限失败: %w", err)
		}
		if !isOperatorSuperAdmin {
			if containsSuperAdminRole {
				return errs.Forbidden("仅超级管理员可分配超级管理员角色")
			}
			return errs.Forbidden("仅超级管理员可调整超级管理员角色")
		}
	}
	if targetHadSuperAdmin && !containsSuperAdminRole {
		return errs.Forbidden("不能移除超级管理员角色")
	}

	if len(req.RoleIDs) == 0 {
		if err := s.employeeRoleRepo.ReplaceByEmployeeID(ctx, req.EmployeeID, nil); err != nil {
			s.log.Error(ctx, "清空员工角色失败", logger.Error(err), logger.Uint("employeeId", req.EmployeeID))
			return fmt.Errorf("分配角色失败: %w", err)
		}
		s.permSvc.InvalidateEmployeePermCache(ctx, req.EmployeeID)
		return nil
	}

	items := make([]*model.EmployeeRole, 0, len(req.RoleIDs))
	for _, rid := range req.RoleIDs {
		items = append(items, &model.EmployeeRole{EmployeeID: req.EmployeeID, RoleID: rid})
	}
	if err := s.employeeRoleRepo.ReplaceByEmployeeID(ctx, req.EmployeeID, items); err != nil {
		return fmt.Errorf("分配角色失败: %w", err)
	}
	s.permSvc.InvalidateEmployeePermCache(ctx, req.EmployeeID)
	return nil
}

func (s *EmployeeService) departmentSubtreeIDs(ctx context.Context, departmentID uint) ([]uint, error) {
	dept, err := s.deptRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, fmt.Errorf("调用 GetByID 失败: %w", err)
	}
	list, err := s.deptRepo.ListByPathPrefix(ctx, dept.Path)
	if err != nil {
		return nil, fmt.Errorf("调用 ListByPathPrefix 失败: %w", err)
	}
	ids := make([]uint, 0, len(list))
	for _, d := range list {
		ids = append(ids, d.ID)
	}
	return ids, nil
}
