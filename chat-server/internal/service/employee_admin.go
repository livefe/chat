package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/auth"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var employeeLoginStatusText = map[uint8]string{
	0:                      "失败",
	model.SMSStatusSuccess: "成功",
}

func (s *EmployeeService) AdminList(ctx context.Context, req *types.EmployeeListRequest) (*types.EmployeeListResponse, error) {
	filter := repository.EmployeeListFilter{
		ID:         req.ID,
		Mobile:     strings.TrimSpace(req.Mobile),
		Name:       strings.TrimSpace(req.Name),
		EmployeeNo: strings.TrimSpace(req.EmployeeNo),
		RoleID:     req.RoleID,
		Status:     req.Status,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Page:       req.Page,
		PageSize:   req.PageSize,
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
	employeeIDs := make([]uint, 0, len(employees))
	for _, e := range employees {
		employeeIDs = append(employeeIDs, e.ID)
	}
	lastLogins, err := s.loginRecordRepo.LatestSuccessByEmployeeIDs(ctx, employeeIDs)
	if err != nil {
		s.log.Error(ctx, "查询员工最近登录失败", logger.Error(err))
		return nil, fmt.Errorf("查询员工列表失败: %w", err)
	}

	avatars := employeeAvatarURLs(ctx, s.mediaSvc, employees)
	list := make([]types.EmployeeItem, 0, len(employees))
	for _, e := range employees {
		list = append(list, toEmployeeItem(e, rolesByEmployee[e.ID], departmentNameByID[e.DepartmentID], avatars[e.ID], lastLogins[e.ID]))
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
				roles = append(roles, toEmployeeRoleBrief(r))
			}
		}
		sort.SliceStable(roles, func(i, j int) bool {
			return roles[i].ID < roles[j].ID
		})
		out[e.ID] = roles
	}
	return out, nil
}

func toEmployeeRoleBrief(r *model.Role) types.EmployeeRoleBrief {
	return types.EmployeeRoleBrief{ID: r.ID, Name: r.Name, Code: r.Code, Status: r.Status}
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

func (s *EmployeeService) AdminUpdate(ctx context.Context, operatorID uint, req *types.UpdateEmployeeRequest) error {
	employee, err := s.getEmployee(ctx, req.ID)
	if err != nil {
		return err
	}

	mobileChanged := req.Mobile != employee.Mobile
	if err := s.assertCanUpdateEmployee(ctx, operatorID, employee.ID, mobileChanged); err != nil {
		return err
	}
	if mobileChanged {
		other, err := s.employeeRepo.GetByMobile(ctx, req.Mobile)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("系统错误，请稍后重试: %w", err)
		}
		if other != nil && other.ID != employee.ID {
			return errs.Conflict("手机号已被其他员工使用")
		}
	}

	deptID := employee.DepartmentID
	if req.DepartmentID != employee.DepartmentID {
		deptID, _, err = s.resolveDepartment(ctx, req.DepartmentID)
		if err != nil {
			return err
		}
	}
	employee.Mobile = req.Mobile
	employee.Name = req.Name
	employee.EmployeeNo = req.EmployeeNo
	employee.DepartmentID = deptID
	employee.Remark = req.Remark
	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		s.log.Error(ctx, "更新员工失败", logger.Error(err))
		return fmt.Errorf("更新员工失败: %w", err)
	}
	if mobileChanged && operatorID != employee.ID {
		s.authSvc.Revoke(ctx, auth.SubjectTypeEmployee, employee.ID)
	}
	return nil
}

func (s *EmployeeService) assertCanUpdateEmployee(ctx context.Context, operatorID, targetID uint, mobileChanged bool) error {
	if operatorID == targetID {
		return nil
	}
	if mobileChanged {
		return s.permSvc.AssertCanManageEmployee(ctx, operatorID, targetID)
	}
	isOperatorSuper, err := s.permSvc.IsSuperAdmin(ctx, operatorID)
	if err != nil {
		return fmt.Errorf("校验操作者权限失败: %w", err)
	}
	if isOperatorSuper {
		return nil
	}
	isTargetSuper, err := s.permSvc.IsSuperAdmin(ctx, targetID)
	if err != nil {
		return fmt.Errorf("校验权限失败: %w", err)
	}
	if isTargetSuper {
		return errs.Forbidden("仅超级管理员可修改超级管理员信息")
	}
	return nil
}

func (s *EmployeeService) AdminUpdateStatus(ctx context.Context, operatorID, id uint, status uint8) error {
	if operatorID == id && status != model.EmployeeStatusEnabled {
		return errs.Forbidden("不能禁用自己的账号")
	}
	employee, err := s.getEmployee(ctx, id)
	if err != nil {
		return err
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

func (s *EmployeeService) AdminRoleOptions(ctx context.Context, operatorID, employeeID uint) (*types.EmployeeRoleOptionsResponse, error) {
	var assignedIDs []uint
	if employeeID > 0 {
		if _, err := s.getEmployee(ctx, employeeID); err != nil {
			return nil, err
		}
		ids, err := s.employeeRoleRepo.ListRoleIDsByEmployeeID(ctx, employeeID)
		if err != nil {
			return nil, fmt.Errorf("获取员工角色失败: %w", err)
		}
		assignedIDs = ids
	}
	enabled := model.RoleStatusEnabled
	roles, _, err := s.roleRepo.List(ctx, repository.RoleListFilter{Status: &enabled, Page: 1, PageSize: 1000})
	if err != nil {
		return nil, fmt.Errorf("获取角色失败: %w", err)
	}
	seen := make(map[uint]struct{}, len(roles))
	for _, r := range roles {
		seen[r.ID] = struct{}{}
	}
	missing := make([]uint, 0)
	for _, id := range assignedIDs {
		if _, ok := seen[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		extra, err := s.roleRepo.GetByIDs(ctx, missing)
		if err != nil {
			return nil, fmt.Errorf("获取角色失败: %w", err)
		}
		roles = append(roles, extra...)
	}

	out := &types.EmployeeRoleOptionsResponse{
		Options:     make([]types.EmployeeRoleBrief, 0, len(roles)),
		AssignedIDs: assignedIDs,
	}
	if out.AssignedIDs == nil {
		out.AssignedIDs = []uint{}
	}
	for _, r := range roles {
		out.Options = append(out.Options, toEmployeeRoleBrief(r))
		if r.Code == types.SuperAdminRoleCode && containsUint(assignedIDs, r.ID) {
			out.TargetIsSuper = true
		}
	}
	sort.SliceStable(out.Options, func(i, j int) bool {
		return out.Options[i].ID < out.Options[j].ID
	})
	out.OperatorIsSuper, err = s.permSvc.IsSuperAdmin(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("校验操作者权限失败: %w", err)
	}
	return out, nil
}

func (s *EmployeeService) AdminAssignRoles(ctx context.Context, operatorID uint, req *types.AssignEmployeeRolesRequest) error {
	if operatorID == req.EmployeeID {
		return errs.Forbidden("不能修改自己的角色")
	}
	if _, err := s.getEmployee(ctx, req.EmployeeID); err != nil {
		return err
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

	roleIDs := uniqueUints(req.RoleIDs)
	containsSuperAdminRole := false
	if len(roleIDs) > 0 {
		roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
		if err != nil {
			return fmt.Errorf("查询角色失败: %w", err)
		}
		if len(roles) != len(roleIDs) {
			return errs.NotFound("部分角色不存在")
		}
		for _, role := range roles {
			if role.Status != model.RoleStatusEnabled && !containsUint(existingRoleIDs, role.ID) {
				return errs.Invalid(fmt.Sprintf("角色「%s」已禁用，不能分配", role.Name))
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
			if containsSuperAdminRole && !targetHadSuperAdmin {
				return errs.Forbidden("仅超级管理员可分配超级管理员角色")
			}
			return errs.Forbidden("仅超级管理员可调整超级管理员的角色")
		}
	}
	if targetHadSuperAdmin && !containsSuperAdminRole {
		return errs.Forbidden("不能移除超级管理员角色")
	}
	if err := s.permSvc.AssertCanManageEmployee(ctx, operatorID, req.EmployeeID); err != nil {
		return err
	}
	addedRoleIDs := make([]uint, 0, len(roleIDs))
	for _, rid := range roleIDs {
		if !containsUint(existingRoleIDs, rid) {
			addedRoleIDs = append(addedRoleIDs, rid)
		}
	}
	if err := s.permSvc.AssertCanGrantRoles(ctx, operatorID, addedRoleIDs); err != nil {
		return err
	}

	items := make([]*model.EmployeeRole, 0, len(roleIDs))
	for _, rid := range roleIDs {
		items = append(items, &model.EmployeeRole{EmployeeID: req.EmployeeID, RoleID: rid})
	}
	if err := s.employeeRoleRepo.ReplaceByEmployeeID(ctx, req.EmployeeID, items); err != nil {
		s.log.Error(ctx, "分配员工角色失败", logger.Error(err), logger.Uint("employeeId", req.EmployeeID))
		return fmt.Errorf("分配角色失败: %w", err)
	}
	s.permSvc.InvalidateEmployeePermCache(ctx, req.EmployeeID)
	return nil
}

func (s *EmployeeService) AdminLoginRecords(ctx context.Context, req *types.EmployeeLoginRecordListRequest) (*types.EmployeeLoginRecordListResponse, error) {
	employee, err := s.getEmployee(ctx, req.EmployeeID)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.loginRecordRepo.ListByEmployee(ctx, employee.ID, employee.Mobile, req.Page, req.PageSize)
	if err != nil {
		s.log.Error(ctx, "查询员工登录记录失败", logger.Error(err))
		return nil, fmt.Errorf("查询登录记录失败: %w", err)
	}
	list := make([]types.EmployeeLoginRecordItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.EmployeeLoginRecordItem{
			ID:         r.ID,
			LoginType:  r.LoginType,
			IP:         r.IP,
			UserAgent:  r.UserAgent,
			Status:     r.Status,
			StatusText: utils.TextOf(employeeLoginStatusText, r.Status),
			Remark:     r.Remark,
			CreatedAt:  utils.FormatDateTime(r.CreatedAt),
		})
	}
	return &types.EmployeeLoginRecordListResponse{Total: total, List: list}, nil
}

func (s *EmployeeService) getEmployee(ctx context.Context, id uint) (*model.Employee, error) {
	employee, err := s.employeeRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("员工不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	return employee, nil
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

func containsUint(list []uint, v uint) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func uniqueUints(list []uint) []uint {
	seen := make(map[uint]struct{}, len(list))
	out := make([]uint, 0, len(list))
	for _, v := range list {
		if v == 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
