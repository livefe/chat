package service

import (
	"context"
	"fmt"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"
)

var roleStatusText = map[uint8]string{
	model.RoleStatusDisabled: "禁用",
	model.RoleStatusEnabled:  "启用",
}

type RoleService struct {
	log                *logger.Logger
	db                 *database.DB
	roleRepo           repository.RoleRepository
	employeeRoleRepo   repository.EmployeeRoleRepository
	rolePermissionRepo repository.RolePermissionRepository
	permissionRepo     repository.PermissionRepository
	permSvc            *PermissionService
}

func NewRoleService(
	log *logger.Logger,
	db *database.DB,
	roleRepo repository.RoleRepository,
	employeeRoleRepo repository.EmployeeRoleRepository,
	rolePermissionRepo repository.RolePermissionRepository,
	permissionRepo repository.PermissionRepository,
	permSvc *PermissionService,
) *RoleService {
	return &RoleService{
		log:                log,
		db:                 db,
		roleRepo:           roleRepo,
		employeeRoleRepo:   employeeRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
		permSvc:            permSvc,
	}
}

func (s *RoleService) Create(ctx context.Context, req *types.CreateRoleRequest) (uint, error) {
	if existing, _ := s.roleRepo.GetByCode(ctx, req.Code); existing != nil {
		return 0, errs.Conflict("角色编码已存在")
	}
	role := &model.Role{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      req.Status,
	}
	if err := s.roleRepo.Create(ctx, role); err != nil {
		s.log.Error(ctx, "创建角色失败", logger.Error(err))
		return 0, fmt.Errorf("创建角色失败: %w", err)
	}
	return role.ID, nil
}

func (s *RoleService) Update(ctx context.Context, req *types.UpdateRoleRequest) error {
	role, err := s.roleRepo.GetByID(ctx, req.ID)
	if err != nil {
		return errs.NotFound("角色不存在")
	}
	if role.Code == types.SuperAdminRoleCode {
		if req.Code != types.SuperAdminRoleCode {
			return errs.Forbidden("不能修改超级管理员角色编码")
		}
		if req.Status != model.RoleStatusEnabled {
			return errs.Forbidden("不能禁用超级管理员角色")
		}
	}
	if req.Code != role.Code {
		if existing, _ := s.roleRepo.GetByCode(ctx, req.Code); existing != nil && existing.ID != req.ID {
			return errs.Conflict("角色编码已存在")
		}
	}
	statusChanged := role.Status != req.Status
	role.Name = req.Name
	role.Code = req.Code
	role.Description = req.Description
	role.Status = req.Status
	if err := s.roleRepo.Update(ctx, role); err != nil {
		s.log.Error(ctx, "更新角色失败", logger.Error(err))
		return fmt.Errorf("更新角色失败: %w", err)
	}
	if statusChanged {
		s.permSvc.InvalidateAllEmployeePermCache(ctx)
	}
	return nil
}

func (s *RoleService) Delete(ctx context.Context, id uint) error {
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return errs.NotFound("角色不存在")
	}
	if role.Code == types.SuperAdminRoleCode {
		return errs.Forbidden("不能删除超级管理员角色")
	}

	role.Code = fmt.Sprintf("%s__deleted_%d", role.Code, role.ID)
	role.Status = model.RoleStatusDisabled
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.roleRepo.Update(ctx, role); err != nil {
			return err
		}
		if err := s.roleRepo.Delete(ctx, id); err != nil {
			return err
		}
		if err := s.rolePermissionRepo.DeleteByRoleID(ctx, id); err != nil {
			return err
		}
		return s.employeeRoleRepo.DeleteByRoleID(ctx, id)
	}); err != nil {
		s.log.Error(ctx, "删除角色失败", logger.Error(err))
		return fmt.Errorf("删除角色失败: %w", err)
	}
	s.permSvc.InvalidateAllEmployeePermCache(ctx)
	return nil
}

func (s *RoleService) List(ctx context.Context, req *types.RoleListRequest) (*types.RoleListResponse, error) {
	roles, total, err := s.roleRepo.List(ctx, repository.RoleListFilter{
		Name: strings.TrimSpace(req.Name), Code: strings.TrimSpace(req.Code), Status: req.Status,
		Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询角色列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}
	roleIDs := make([]uint, 0, len(roles))
	for _, r := range roles {
		roleIDs = append(roleIDs, r.ID)
	}
	memberCounts, err := s.employeeRoleRepo.CountGroupByRoleIDs(ctx, roleIDs)
	if err != nil {
		s.log.Error(ctx, "统计角色成员失败", logger.Error(err))
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}

	list := make([]types.RoleItem, 0, len(roles))
	for _, r := range roles {
		item := toRoleItem(r)
		item.MemberCount = memberCounts[r.ID]
		list = append(list, item)
	}
	return &types.RoleListResponse{Total: total, List: list}, nil
}

func (s *RoleService) ListPermissions(ctx context.Context, roleID uint) ([]types.PermissionItem, error) {
	ids, err := s.rolePermissionRepo.GetPermissionIDsByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("获取角色权限失败: %w", err)
	}
	if len(ids) == 0 {
		return []types.PermissionItem{}, nil
	}

	perms, err := s.permissionRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("获取权限详情失败: %w", err)
	}

	items := make([]types.PermissionItem, 0, len(perms))
	for _, p := range perms {
		if !isCapabilityType(p.Type) {
			continue
		}
		items = append(items, toPermissionItem(p))
	}
	return items, nil
}

func (s *RoleService) AssignPermissions(ctx context.Context, operatorID uint, req *types.AssignPermissionsRequest) error {
	if _, err := s.roleRepo.GetByID(ctx, req.RoleID); err != nil {
		return errs.NotFound("角色不存在")
	}

	requested := uniqueUints(req.PermissionIDs)
	if len(requested) > 0 {
		perms, err := s.permissionRepo.GetByIDs(ctx, requested)
		if err != nil {
			return fmt.Errorf("查询权限失败: %w", err)
		}
		if len(perms) != len(requested) {
			return errs.NotFound("部分权限不存在")
		}
		for _, p := range perms {
			if !isCapabilityType(p.Type) {
				return errs.Invalid("接口权限不可分配，请只勾选菜单或按钮权限")
			}
		}
	}

	finalIDs, err := s.scopePermissionIDs(ctx, operatorID, req.RoleID, requested)
	if err != nil {
		return err
	}

	items := make([]*model.RolePermission, 0, len(finalIDs))
	for _, pid := range finalIDs {
		items = append(items, &model.RolePermission{RoleID: req.RoleID, PermissionID: pid})
	}
	if err := s.rolePermissionRepo.ReplaceByRoleID(ctx, req.RoleID, items); err != nil {
		s.log.Error(ctx, "分配角色权限失败", logger.Error(err), logger.Uint("roleId", req.RoleID))
		return fmt.Errorf("分配权限失败: %w", err)
	}
	s.permSvc.InvalidateAllEmployeePermCache(ctx)
	return nil
}

// scopePermissionIDs 只在操作者自身权限范围内调整；角色原有的、超出操作者范围的权限原样保留。
func (s *RoleService) scopePermissionIDs(ctx context.Context, operatorID, roleID uint, requested []uint) ([]uint, error) {
	isSuper, err := s.permSvc.IsSuperAdmin(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("校验操作者权限失败: %w", err)
	}
	if isSuper {
		return requested, nil
	}
	allowed, err := s.permSvc.EnabledPermissions(ctx, operatorID)
	if err != nil {
		return nil, fmt.Errorf("校验操作者权限失败: %w", err)
	}
	inScope := make(map[uint]struct{}, len(allowed))
	for _, p := range allowed {
		inScope[p.ID] = struct{}{}
	}
	existing, err := s.rolePermissionRepo.GetPermissionIDsByRoleID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("获取角色权限失败: %w", err)
	}

	out := make([]uint, 0, len(requested)+len(existing))
	for _, id := range requested {
		if _, ok := inScope[id]; ok {
			out = append(out, id)
		} else if !containsUint(existing, id) {
			return nil, errs.Forbidden("不能分配超出自身权限范围的权限")
		}
	}
	for _, id := range existing {
		if _, ok := inScope[id]; !ok {
			out = append(out, id)
		}
	}
	return out, nil
}

func toRoleItem(r *model.Role) types.RoleItem {
	return types.RoleItem{
		ID:          r.ID,
		Name:        r.Name,
		Code:        r.Code,
		Description: r.Description,
		Status:      r.Status,
		StatusText:  utils.TextOf(roleStatusText, r.Status),
		CreatedAt:   utils.FormatDateTime(r.CreatedAt),
		UpdatedAt:   utils.FormatDateTime(r.UpdatedAt),
	}
}
