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
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

var departmentStatusText = map[uint8]string{
	model.DepartmentStatusDisabled: "禁用",
	model.DepartmentStatusEnabled:  "启用",
}

type DepartmentService struct {
	log          *logger.Logger
	employeeRepo repository.EmployeeRepository
	deptRepo     repository.DepartmentRepository
}

func NewDepartmentService(
	log *logger.Logger,
	employeeRepo repository.EmployeeRepository,
	deptRepo repository.DepartmentRepository,
) *DepartmentService {
	return &DepartmentService{
		log:          log,
		employeeRepo: employeeRepo,
		deptRepo:     deptRepo,
	}
}

func (s *DepartmentService) List(ctx context.Context, req *types.DepartmentListRequest) (*types.DepartmentTreeResponse, error) {
	all, err := s.deptRepo.List(ctx, repository.DepartmentListFilter{})
	if err != nil {
		s.log.Error(ctx, "查询部门列表失败", logger.Error(err))
		return nil, fmt.Errorf("查询部门列表失败: %w", err)
	}
	ids := make([]uint, 0, len(all))
	for _, d := range all {
		ids = append(ids, d.ID)
	}
	countMap, err := s.employeeRepo.CountGroupByDepartmentIDs(ctx, ids)
	if err != nil {
		s.log.Error(ctx, "统计部门员工数失败", logger.Error(err))
		return nil, fmt.Errorf("查询部门列表失败: %w", err)
	}
	nameByID := make(map[uint]string, len(all))
	for _, d := range all {
		nameByID[d.ID] = d.Name
	}

	keep := make(map[uint]struct{})
	nameQ := strings.TrimSpace(req.Name)
	for _, d := range all {
		if nameQ != "" && !strings.Contains(d.Name, nameQ) {
			continue
		}
		if req.Status != nil && d.Status != *req.Status {
			continue
		}
		keep[d.ID] = struct{}{}
		for _, pid := range ancestorIDs(d.Path, d.ID) {
			keep[pid] = struct{}{}
		}
	}
	if nameQ == "" && req.Status == nil {
		for _, d := range all {
			keep[d.ID] = struct{}{}
		}
	}

	nodes := make(map[uint]*types.DepartmentTreeItem, len(all))
	ordered := make([]*model.Department, 0, len(all))
	for _, d := range all {
		if _, ok := keep[d.ID]; !ok {
			continue
		}
		item := toDepartmentItem(d, nameByID[d.ParentID])
		item.EmployeeCount = countMap[d.ID]
		nodes[d.ID] = &types.DepartmentTreeItem{DepartmentItem: item, Children: []types.DepartmentTreeItem{}}
		ordered = append(ordered, d)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Level != ordered[j].Level {
			return ordered[i].Level > ordered[j].Level
		}
		if ordered[i].Sort != ordered[j].Sort {
			return ordered[i].Sort < ordered[j].Sort
		}
		return ordered[i].ID < ordered[j].ID
	})

	for _, d := range ordered {
		node := nodes[d.ID]
		if node == nil || d.ParentID == 0 {
			continue
		}
		if parent := nodes[d.ParentID]; parent != nil {
			parent.Children = append(parent.Children, *node)
		}
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Level != ordered[j].Level {
			return ordered[i].Level < ordered[j].Level
		}
		if ordered[i].Sort != ordered[j].Sort {
			return ordered[i].Sort < ordered[j].Sort
		}
		return ordered[i].ID < ordered[j].ID
	})
	roots := make([]types.DepartmentTreeItem, 0)
	for _, d := range ordered {
		if d.ParentID > 0 && nodes[d.ParentID] != nil {
			continue
		}
		if node := nodes[d.ID]; node != nil {
			roots = append(roots, *node)
		}
	}
	return &types.DepartmentTreeResponse{List: roots}, nil
}

func ancestorIDs(path string, selfID uint) []uint {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	out := make([]uint, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		var id uint
		if _, err := fmt.Sscanf(p, "%d", &id); err != nil || id == 0 || id == selfID {
			continue
		}
		out = append(out, id)
	}
	return out
}

func (s *DepartmentService) ListOptions(ctx context.Context) ([]types.DepartmentOption, error) {
	list, err := s.deptRepo.ListEnabled(ctx)
	if err != nil {
		s.log.Error(ctx, "查询部门选项失败", logger.Error(err))
		return nil, fmt.Errorf("查询部门选项失败: %w", err)
	}
	nodes := make(map[uint]*types.DepartmentOption, len(list))
	for _, d := range list {
		nodes[d.ID] = &types.DepartmentOption{ID: d.ID, Name: d.Name, Children: []types.DepartmentOption{}}
	}
	ordered := append([]*model.Department(nil), list...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Level != ordered[j].Level {
			return ordered[i].Level > ordered[j].Level
		}
		if ordered[i].Sort != ordered[j].Sort {
			return ordered[i].Sort < ordered[j].Sort
		}
		return ordered[i].ID < ordered[j].ID
	})
	for _, d := range ordered {
		node := nodes[d.ID]
		if d.ParentID == 0 {
			continue
		}
		if parent := nodes[d.ParentID]; parent != nil {
			parent.Children = append(parent.Children, *node)
		}
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Level != ordered[j].Level {
			return ordered[i].Level < ordered[j].Level
		}
		if ordered[i].Sort != ordered[j].Sort {
			return ordered[i].Sort < ordered[j].Sort
		}
		return ordered[i].ID < ordered[j].ID
	})
	roots := make([]types.DepartmentOption, 0)
	for _, d := range ordered {
		if d.ParentID > 0 && nodes[d.ParentID] != nil {
			continue
		}
		if node := nodes[d.ID]; node != nil {
			roots = append(roots, *node)
		}
	}
	return roots, nil
}

func (s *DepartmentService) Create(ctx context.Context, req *types.CreateDepartmentRequest) (uint, error) {
	level := uint8(1)
	parentPath := ""
	if req.ParentID > 0 {
		parent, err := s.deptRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, errs.NotFound("上级部门不存在")
			}
			return 0, fmt.Errorf("查询上级部门失败: %w", err)
		}
		if parent.Status != model.DepartmentStatusEnabled {
			return 0, errs.Invalid("上级部门已禁用")
		}
		if parent.Level >= types.DepartmentMaxLevel {
			return 0, errs.Invalid(fmt.Sprintf("部门最多 %d 级", types.DepartmentMaxLevel))
		}
		level = parent.Level + 1
		parentPath = parent.Path
	}
	dept := &model.Department{
		ParentID: req.ParentID,
		Name:     req.Name,
		Level:    level,
		Status:   req.Status,
		Sort:     req.Sort,
		Remark:   req.Remark,
	}
	if err := s.deptRepo.Create(ctx, dept); err != nil {
		s.log.Error(ctx, "创建部门失败", logger.Error(err))
		return 0, fmt.Errorf("创建部门失败: %w", err)
	}
	if parentPath == "" {
		dept.Path = fmt.Sprintf("/%d/", dept.ID)
	} else {
		dept.Path = parentPath + fmt.Sprintf("%d/", dept.ID)
	}
	if err := s.deptRepo.Update(ctx, dept); err != nil {
		s.log.Error(ctx, "更新部门路径失败", logger.Error(err))
		return 0, fmt.Errorf("创建部门失败: %w", err)
	}
	return dept.ID, nil
}

func (s *DepartmentService) Update(ctx context.Context, req *types.UpdateDepartmentRequest) error {
	dept, err := s.deptRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("部门不存在")
		}
		return fmt.Errorf("查询部门失败: %w", err)
	}
	if req.Status == model.DepartmentStatusDisabled && dept.Status != model.DepartmentStatusDisabled {
		if err := s.assertDepartmentDisableable(ctx, dept.ID); err != nil {
			return err
		}
	}
	oldParentID := dept.ParentID
	oldPath := dept.Path
	if req.ParentID != oldParentID {
		if err := s.assertCanMove(ctx, dept, req.ParentID); err != nil {
			return err
		}
	}

	dept.Name = req.Name
	dept.Status = req.Status
	dept.Sort = req.Sort
	dept.Remark = req.Remark
	dept.ParentID = req.ParentID

	if req.ParentID != oldParentID {
		level := uint8(1)
		newPath := ""
		if req.ParentID > 0 {
			parent, err := s.deptRepo.GetByID(ctx, req.ParentID)
			if err != nil {
				return fmt.Errorf("查询上级部门失败: %w", err)
			}
			level = parent.Level + 1
			newPath = parent.Path + fmt.Sprintf("%d/", dept.ID)
		} else {
			newPath = fmt.Sprintf("/%d/", dept.ID)
		}
		dept.Level = level
		dept.Path = newPath
	}

	if err := s.deptRepo.Update(ctx, dept); err != nil {
		s.log.Error(ctx, "更新部门失败", logger.Error(err))
		return fmt.Errorf("更新部门失败: %w", err)
	}

	if req.ParentID != oldParentID && oldPath != "" {
		if err := s.rebuildDescendantPaths(ctx, oldPath, dept.Path, dept.Level); err != nil {
			s.log.Error(ctx, "重建子部门路径失败", logger.Error(err))
			return fmt.Errorf("更新部门失败: %w", err)
		}
	}
	return nil
}

func (s *DepartmentService) assertCanMove(ctx context.Context, dept *model.Department, newParentID uint) error {
	if newParentID == 0 {
		return nil
	}
	if newParentID == dept.ID {
		return errs.Invalid("上级部门不能是自己")
	}
	parent, err := s.deptRepo.GetByID(ctx, newParentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("上级部门不存在")
		}
		return fmt.Errorf("查询上级部门失败: %w", err)
	}
	if parent.Status != model.DepartmentStatusEnabled {
		return errs.Invalid("上级部门已禁用")
	}
	if dept.Path != "" && strings.HasPrefix(parent.Path, dept.Path) {
		return errs.Invalid("不能移动到自己的下级部门下")
	}
	subtree, err := s.deptRepo.ListByPathPrefix(ctx, dept.Path)
	if err != nil {
		return fmt.Errorf("查询子部门失败: %w", err)
	}
	maxRelative := uint8(0)
	for _, child := range subtree {
		if child.ID == dept.ID {
			continue
		}
		rel := child.Level - dept.Level
		if rel > maxRelative {
			maxRelative = rel
		}
	}
	if parent.Level+1+maxRelative > types.DepartmentMaxLevel {
		return errs.Invalid(fmt.Sprintf("移动后将超过 %d 级限制", types.DepartmentMaxLevel))
	}
	return nil
}

func (s *DepartmentService) rebuildDescendantPaths(ctx context.Context, oldPath, newPath string, newRootLevel uint8) error {
	children, err := s.deptRepo.ListByPathPrefix(ctx, oldPath)
	if err != nil {
		return fmt.Errorf("调用 ListByPathPrefix 失败: %w", err)
	}
	oldLevel := uint8(strings.Count(strings.Trim(oldPath, "/"), "/") + 1)
	for _, child := range children {
		if child.Path == oldPath || !strings.HasPrefix(child.Path, oldPath) {
			continue
		}
		suffix := strings.TrimPrefix(child.Path, oldPath)
		child.Path = newPath + suffix
		rel := child.Level - oldLevel
		child.Level = newRootLevel + rel
		if err := s.deptRepo.Update(ctx, child); err != nil {
			return err
		}
	}
	return nil
}

func (s *DepartmentService) UpdateStatus(ctx context.Context, id uint, status uint8) error {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("部门不存在")
		}
		return fmt.Errorf("查询部门失败: %w", err)
	}
	if status == model.DepartmentStatusDisabled {
		if err := s.assertDepartmentDisableable(ctx, id); err != nil {
			return err
		}
	}
	dept.Status = status
	if err := s.deptRepo.Update(ctx, dept); err != nil {
		s.log.Error(ctx, "更新部门状态失败", logger.Error(err))
		return fmt.Errorf("更新部门状态失败: %w", err)
	}
	return nil
}

func (s *DepartmentService) assertDepartmentDisableable(ctx context.Context, id uint) error {
	count, err := s.employeeRepo.CountByDepartmentID(ctx, id)
	if err != nil {
		s.log.Error(ctx, "统计部门员工数失败", logger.Error(err))
		return fmt.Errorf("查询部门失败: %w", err)
	}
	if count > 0 {
		return errs.Conflict(fmt.Sprintf("该部门下仍有 %d 名员工，无法禁用", count))
	}
	childCount, err := s.deptRepo.CountChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("查询部门失败: %w", err)
	}
	if childCount > 0 {
		return errs.Conflict("请先处理下级部门后再禁用")
	}
	return nil
}

func toDepartmentItem(d *model.Department, parentName string) types.DepartmentItem {
	return types.DepartmentItem{
		ID:         d.ID,
		ParentID:   d.ParentID,
		ParentName: parentName,
		Name:       d.Name,
		Path:       d.Path,
		Level:      d.Level,
		Status:     d.Status,
		StatusText: utils.TextOf(departmentStatusText, d.Status),
		Sort:       d.Sort,
		Remark:     d.Remark,
		CreatedAt:  utils.FormatDateTime(d.CreatedAt),
		UpdatedAt:  utils.FormatDateTime(d.UpdatedAt),
	}
}
