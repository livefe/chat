package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *PermissionService) IsSuperAdmin(ctx context.Context, employeeID uint) (bool, error) {
	roleIDs, err := s.employeeRoleRepo.ListRoleIDsByEmployeeID(ctx, employeeID)
	if err != nil || len(roleIDs) == 0 {
		return false, err
	}
	roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
	if err != nil {
		return false, fmt.Errorf("调用 GetByIDs 失败: %w", err)
	}
	for _, r := range roles {
		if r.Code == types.SuperAdminRoleCode && r.Status == model.RoleStatusEnabled {
			return true, nil
		}
	}
	return false, nil
}

func (s *PermissionService) GetEmployeePermissions(ctx context.Context, employeeID uint) ([]types.PermissionItem, error) {
	perms, err := s.EnabledPermissions(ctx, employeeID)
	if err != nil {
		s.log.Error(ctx, "查询员工权限失败", logger.Error(err), logger.Uint("employeeId", employeeID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	items := make([]types.PermissionItem, 0, len(perms))
	for _, p := range perms {
		items = append(items, toPermissionItem(p))
	}
	return items, nil
}

func (s *PermissionService) CheckPermission(ctx context.Context, employeeID uint, path, method string) (bool, error) {
	isSuper, perms, err := s.EmployeeAccess(ctx, employeeID)
	if err != nil {
		return false, err
	}
	if isSuper {
		return true, nil
	}

	reqPath := strings.TrimSuffix(strings.TrimSpace(path), "/")
	reqMethod := strings.ToUpper(strings.TrimSpace(method))

	apiDef, err := s.permissionRepo.GetAPIByPathAndMethod(ctx, reqPath, reqMethod)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if apiDef == nil {
		return false, nil
	}

	codes := make(map[string]struct{}, len(perms))
	parentIDs := make(map[uint]struct{}, len(perms))
	for _, p := range perms {
		if p == nil || p.Code == "" {
			continue
		}
		codes[p.Code] = struct{}{}
		if p.Type == model.PermissionTypeMenu {
			parentIDs[p.ID] = struct{}{}
		}
	}

	if _, ok := codes[apiDef.Code]; ok {
		return true, nil
	}

	if reqMethod == "GET" && apiDef.ParentID > 0 {
		if _, ok := parentIDs[apiDef.ParentID]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (s *PermissionService) HasPermissionCode(ctx context.Context, employeeID uint, code string) (bool, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, nil
	}
	isSuper, perms, err := s.EmployeeAccess(ctx, employeeID)
	if err != nil {
		return false, err
	}
	if isSuper {
		return true, nil
	}
	for _, p := range perms {
		if p != nil && p.Code == code {
			return true, nil
		}
	}
	return false, nil
}

func (s *PermissionService) EnabledPermissions(ctx context.Context, employeeID uint) ([]*model.Permission, error) {
	_, perms, err := s.EmployeeAccess(ctx, employeeID)
	return perms, err
}

func (s *PermissionService) EmployeeAccess(ctx context.Context, employeeID uint) (bool, []*model.Permission, error) {
	if cached, ok := s.loadEmployeePermCache(ctx, employeeID); ok {
		return cached.IsSuper, cached.Perms, nil
	}

	isSuper, err := s.IsSuperAdmin(ctx, employeeID)
	if err != nil {
		return false, nil, err
	}

	var perms []*model.Permission
	if isSuper {
		all, err := s.permissionRepo.ListAll(ctx)
		if err != nil {
			return false, nil, err
		}
		perms = filterEnabled(all)
	} else {
		roleIDs, err := s.employeeRoleRepo.ListRoleIDsByEmployeeID(ctx, employeeID)
		if err != nil {
			return false, nil, err
		}
		if len(roleIDs) == 0 {
			s.storeEmployeePermCache(ctx, employeeID, false, nil)
			return false, nil, nil
		}

		roles, err := s.roleRepo.GetByIDs(ctx, roleIDs)
		if err != nil {
			return false, nil, err
		}
		enabledRoleIDs := make([]uint, 0, len(roles))
		for _, r := range roles {
			if r.Status == model.RoleStatusEnabled {
				enabledRoleIDs = append(enabledRoleIDs, r.ID)
			}
		}
		if len(enabledRoleIDs) == 0 {
			s.storeEmployeePermCache(ctx, employeeID, false, nil)
			return false, nil, nil
		}

		permIDs, err := s.rolePermissionRepo.GetPermissionIDsByRoleIDs(ctx, enabledRoleIDs)
		if err != nil {
			return false, nil, err
		}
		if len(permIDs) == 0 {
			s.storeEmployeePermCache(ctx, employeeID, false, nil)
			return false, nil, nil
		}

		all, err := s.permissionRepo.GetByIDs(ctx, permIDs)
		if err != nil {
			return false, nil, err
		}
		perms = filterEnabled(all)
	}

	perms = filterCapability(perms)
	s.storeEmployeePermCache(ctx, employeeID, isSuper, perms)
	return isSuper, perms, nil
}

func filterEnabled(perms []*model.Permission) []*model.Permission {
	out := make([]*model.Permission, 0, len(perms))
	for _, p := range perms {
		if p.Status == model.PermissionStatusEnabled {
			out = append(out, p)
		}
	}
	return out
}

func filterCapability(perms []*model.Permission) []*model.Permission {
	out := make([]*model.Permission, 0, len(perms))
	for _, p := range perms {
		if p != nil && isCapabilityType(p.Type) {
			out = append(out, p)
		}
	}
	return out
}
