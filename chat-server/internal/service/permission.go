package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/database"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"
)

type PermissionService struct {
	log                *logger.Logger
	cache              *cache.Cache
	db                 *database.DB
	permissionRepo     repository.PermissionRepository
	roleRepo           repository.RoleRepository
	employeeRoleRepo   repository.EmployeeRoleRepository
	rolePermissionRepo repository.RolePermissionRepository
}

var (
	permissionStatusText = map[uint8]string{
		model.PermissionStatusDisabled: "禁用",
		model.PermissionStatusEnabled:  "启用",
	}
	permissionTypeText = map[uint8]string{
		model.PermissionTypeMenu:   "菜单",
		model.PermissionTypeButton: "按钮",
		model.PermissionTypeAPI:    "接口",
	}
)

func isCapabilityType(t uint8) bool {
	return t == model.PermissionTypeMenu || t == model.PermissionTypeButton
}

func NewPermissionService(
	log *logger.Logger,
	c *cache.Cache,
	db *database.DB,
	permissionRepo repository.PermissionRepository,
	roleRepo repository.RoleRepository,
	employeeRoleRepo repository.EmployeeRoleRepository,
	rolePermissionRepo repository.RolePermissionRepository,
) *PermissionService {
	return &PermissionService{
		log: log, cache: c, db: db, permissionRepo: permissionRepo,
		roleRepo: roleRepo, employeeRoleRepo: employeeRoleRepo, rolePermissionRepo: rolePermissionRepo,
	}
}

func (s *PermissionService) AdminCreate(ctx context.Context, req *types.CreatePermissionRequest) (uint, error) {
	if err := validateAPIPermission(req.Type, req.Path, req.Method); err != nil {
		return 0, err
	}
	if err := s.ensureCodeAvailable(ctx, req.Type, req.Code, req.Path, req.Method, 0); err != nil {
		return 0, err
	}
	if req.ParentID > 0 {
		if _, err := s.permissionRepo.GetByID(ctx, req.ParentID); err != nil {
			return 0, errs.NotFound("父级权限不存在")
		}
	}
	p := &model.Permission{
		Name: req.Name, Code: req.Code, Description: req.Description,
		Type: req.Type, ParentID: req.ParentID,
		Path: req.Path, Method: req.Method,
		Sort: req.Sort, Status: req.Status,
		Icon: req.Icon, Component: req.Component, HideInMenu: req.HideInMenu,
	}
	if err := s.permissionRepo.Create(ctx, p); err != nil {
		s.log.Error(ctx, "创建权限失败", logger.Error(err))
		return 0, fmt.Errorf("创建权限失败: %w", err)
	}
	s.invalidateAllEmployeePermCache(ctx)
	return p.ID, nil
}

func (s *PermissionService) AdminUpdate(ctx context.Context, req *types.UpdatePermissionRequest) error {
	if err := validateAPIPermission(req.Type, req.Path, req.Method); err != nil {
		return err
	}
	p, err := s.permissionRepo.GetByID(ctx, req.ID)
	if err != nil {
		return errs.NotFound("权限不存在")
	}
	if err := s.ensureCodeAvailable(ctx, req.Type, req.Code, req.Path, req.Method, req.ID); err != nil {
		return err
	}
	if req.ParentID > 0 && req.ParentID != p.ID {
		if _, err := s.permissionRepo.GetByID(ctx, req.ParentID); err != nil {
			return errs.NotFound("父级权限不存在")
		}
	}
	p.Name = req.Name
	p.Code = req.Code
	p.Description = req.Description
	p.Type = req.Type
	p.ParentID = req.ParentID
	if p.Type == model.PermissionTypeAPI && req.Type != model.PermissionTypeAPI {
		return errs.Invalid("接口权限类型不可更改")
	}

	if p.Type == model.PermissionTypeAPI || req.Type == model.PermissionTypeAPI {
		if p.Type == model.PermissionTypeAPI && (req.Path != p.Path || strings.ToUpper(strings.TrimSpace(req.Method)) != strings.ToUpper(strings.TrimSpace(p.Method))) {
			return errs.Invalid("接口权限的路径与方法创建后不可修改，请新建权限")
		}
		if p.Path == "" {
			p.Path = req.Path
			p.Method = strings.ToUpper(strings.TrimSpace(req.Method))
		}
	} else {
		p.Path = req.Path
		p.Method = req.Method
	}
	p.Sort = req.Sort
	p.Status = req.Status
	p.Icon = req.Icon
	p.Component = req.Component
	p.HideInMenu = req.HideInMenu
	if err := s.permissionRepo.Update(ctx, p); err != nil {
		s.log.Error(ctx, "更新权限失败", logger.Error(err))
		return fmt.Errorf("更新权限失败: %w", err)
	}
	s.invalidateAllEmployeePermCache(ctx)
	return nil
}

func (s *PermissionService) AdminDelete(ctx context.Context, id uint) error {
	p, err := s.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return errs.NotFound("权限不存在")
	}

	children, err := s.permissionRepo.ListByParentID(ctx, id)
	if err != nil {
		return fmt.Errorf("检查子权限失败: %w", err)
	}
	if len(children) > 0 {
		return errs.Conflict("存在子权限，无法删除")
	}
	n, err := s.rolePermissionRepo.CountByPermissionID(ctx, id)
	if err != nil {
		return fmt.Errorf("检查权限引用失败: %w", err)
	}
	if n > 0 {
		return errs.Conflict("权限已被角色引用，无法删除")
	}
	p.Code = fmt.Sprintf("%s__deleted_%d", p.Code, p.ID)
	p.Status = model.PermissionStatusDisabled
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if err := s.permissionRepo.Update(ctx, p); err != nil {
			return err
		}
		return s.permissionRepo.Delete(ctx, id)
	}); err != nil {
		s.log.Error(ctx, "删除权限失败", logger.Error(err))
		return fmt.Errorf("删除权限失败: %w", err)
	}
	s.invalidateAllEmployeePermCache(ctx)
	return nil
}

func validateAPIPermission(permType uint8, path, method string) error {
	if permType != model.PermissionTypeAPI {
		return nil
	}
	path = strings.TrimSpace(path)
	method = strings.ToUpper(strings.TrimSpace(method))
	if path == "" || path == "*" {
		return errs.Invalid("接口权限必须指定具体路径，禁止使用通配符")
	}
	if method == "" || method == "*" {
		return errs.Invalid("接口权限必须指定具体 HTTP 方法")
	}
	return nil
}

func (s *PermissionService) ensureCodeAvailable(ctx context.Context, permType uint8, code, path, method string, excludeID uint) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return errs.Invalid("权限编码不能为空")
	}
	if permType == model.PermissionTypeAPI {
		existing, err := s.permissionRepo.GetAPIByPathAndMethod(ctx, path, method)
		if err == nil && existing != nil && existing.ID != excludeID {
			return errs.Conflict("接口路径与方法已存在")
		}
		return nil
	}
	existing, err := s.permissionRepo.GetAssignableByCode(ctx, code)
	if err == nil && existing != nil && existing.ID != excludeID {
		return errs.Conflict("权限编码已存在")
	}
	return nil
}

// AdminTree 管理页返回全部权限（含禁用）；assignableOnly 时只返回可分配的启用菜单与按钮。
func (s *PermissionService) AdminTree(ctx context.Context, assignableOnly bool) ([]*types.PermissionTreeItem, error) {
	perms, err := s.permissionRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询权限失败: %w", err)
	}

	selected := make([]*model.Permission, 0, len(perms))
	for _, p := range perms {
		if assignableOnly && (p.Status != model.PermissionStatusEnabled || !isCapabilityType(p.Type)) {
			continue
		}
		selected = append(selected, p)
	}
	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].Sort != selected[j].Sort {
			return selected[i].Sort < selected[j].Sort
		}
		return selected[i].ID < selected[j].ID
	})

	nodes := make(map[uint]*types.PermissionTreeItem, len(selected))
	for _, p := range selected {
		nodes[p.ID] = &types.PermissionTreeItem{PermissionItem: toPermissionItem(p)}
	}

	tree := make([]*types.PermissionTreeItem, 0)
	for _, p := range selected {
		node := nodes[p.ID]
		if parent, ok := nodes[p.ParentID]; p.ParentID > 0 && ok {
			parent.Children = append(parent.Children, node)
		} else {
			tree = append(tree, node)
		}
	}
	return tree, nil
}

func toPermissionItem(p *model.Permission) types.PermissionItem {
	return types.PermissionItem{
		ID:          p.ID,
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Type:        p.Type,
		TypeText:    utils.TextOf(permissionTypeText, p.Type),
		ParentID:    p.ParentID,
		Path:        p.Path,
		Method:      p.Method,
		Sort:        p.Sort,
		Status:      p.Status,
		StatusText:  utils.TextOf(permissionStatusText, p.Status),
		Icon:        p.Icon,
		Component:   p.Component,
		HideInMenu:  p.HideInMenu,
		CreatedAt:   utils.FormatDateTime(p.CreatedAt),
		UpdatedAt:   utils.FormatDateTime(p.UpdatedAt),
	}
}
