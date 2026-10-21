package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

const maxAuditDetailLen = 2000

type AuditLogInput struct {
	EmployeeID    uint
	RequestPath   string
	RequestMethod string
	RequestBody   []byte
	IP            string
	UserAgent     string
	Success       bool
}

type AuditLogService struct {
	log            *logger.Logger
	auditLogRepo   repository.AuditLogRepository
	employeeRepo   repository.EmployeeRepository
	permissionRepo repository.PermissionRepository
}

func NewAuditLogService(
	log *logger.Logger,
	auditLogRepo repository.AuditLogRepository,
	employeeRepo repository.EmployeeRepository,
	permissionRepo repository.PermissionRepository,
) *AuditLogService {
	return &AuditLogService{
		log:            log,
		auditLogRepo:   auditLogRepo,
		employeeRepo:   employeeRepo,
		permissionRepo: permissionRepo,
	}
}

func (s *AuditLogService) Record(ctx context.Context, in AuditLogInput) {
	module, moduleText, action, actionText, targetType, ok := s.resolveAuditMeta(ctx, in.RequestPath, in.RequestMethod)
	if !ok {
		return
	}

	employee, err := s.employeeRepo.GetByID(ctx, in.EmployeeID)
	if err != nil {
		s.log.Error(ctx, "写入操作审计失败：查询员工失败", logger.Uint("employee_id", in.EmployeeID), logger.Error(err))
		return
	}

	targetID, targetLabel := parseAuditTarget(in.RequestBody)
	detail := sanitizeAuditDetail(in.RequestBody)

	row := &model.AuditLog{
		EmployeeID:     in.EmployeeID,
		EmployeeName:   employee.Name,
		EmployeeMobile: employee.Mobile,
		Module:         module,
		ModuleText:     moduleText,
		Action:         action,
		ActionText:     actionText,
		TargetType:     targetType,
		TargetID:       targetID,
		TargetLabel:    targetLabel,
		RequestPath:    in.RequestPath,
		RequestMethod:  in.RequestMethod,
		Success:        in.Success,
		Detail:         detail,
		IP:             in.IP,
		UserAgent:      in.UserAgent,
	}
	if err := s.auditLogRepo.Create(ctx, row); err != nil {
		s.log.Error(ctx, "写入操作审计失败", logger.Error(err))
	}
}

func (s *AuditLogService) List(ctx context.Context, req *types.AuditLogListRequest) (*types.AuditLogListResponse, error) {
	startTime, endTime, err := parseAuditTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	rows, total, err := s.auditLogRepo.List(ctx, repository.AuditLogListFilter{
		EmployeeID: req.EmployeeID,
		Module:     req.Module,
		Action:     req.Action,
		Keyword:    req.Keyword,
		StartTime:  startTime,
		EndTime:    endTime,
		Page:       req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询操作审计失败", logger.Error(err))
		return nil, fmt.Errorf("查询操作审计失败: %w", err)
	}

	list := make([]types.AuditLogItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toAuditLogItem(row))
	}
	modules, actions, err := s.listAuditOptions(ctx)
	if err != nil {
		s.log.Error(ctx, "查询审计筛选项失败", logger.Error(err))
		return nil, fmt.Errorf("查询操作审计失败: %w", err)
	}
	return &types.AuditLogListResponse{Total: total, List: list, Modules: modules, Actions: actions}, nil
}

func (s *AuditLogService) resolveAuditMeta(
	ctx context.Context,
	path, method string,
) (module, moduleText, action, actionText, targetType string, ok bool) {
	perm, err := s.permissionRepo.GetAPIByPathAndMethod(ctx, path, method)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error(ctx, "查询审计元数据失败", logger.Error(err))
		}
		return "", "", "", "", "", false
	}
	var parent *model.Permission
	if perm.ParentID > 0 {
		parent, err = s.permissionRepo.GetByID(ctx, perm.ParentID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error(ctx, "查询审计父级权限失败", logger.Uint("parent_id", perm.ParentID), logger.Error(err))
		}
	}
	module, moduleText, action, actionText, targetType = auditMetaFromPermission(perm, parent)
	return module, moduleText, action, actionText, targetType, true
}

func auditMetaFromPermission(perm, parent *model.Permission) (module, moduleText, action, actionText, targetType string) {
	actionText = perm.Name
	action = auditActionFromCode(perm.Code)
	targetType = auditModuleFromCode(perm.Code)
	module = targetType
	moduleText = perm.Name

	if parent == nil {
		return module, moduleText, action, actionText, targetType
	}

	module = auditModuleFromMenuCode(parent.Code)
	moduleText = parent.Name
	return module, moduleText, action, actionText, targetType
}

func (s *AuditLogService) listAuditOptions(ctx context.Context) ([]types.AuditModuleOption, []types.AuditActionOption, error) {
	apis, err := s.permissionRepo.ListEnabledAPIs(ctx)
	if err != nil {
		return nil, nil, err
	}

	parentMap, err := s.loadParentPermissions(ctx, apis)
	if err != nil {
		return nil, nil, err
	}

	moduleSeen := make(map[string]string)
	actionSeen := make(map[string]string)
	for _, api := range apis {
		if strings.ToUpper(strings.TrimSpace(api.Method)) != http.MethodPost {
			continue
		}
		module, moduleText, action, actionText, _ := auditMetaFromPermission(api, parentMap[api.ParentID])
		if module != "" {
			moduleSeen[module] = moduleText
		}
		if action != "" {
			actionSeen[action] = actionText
		}
	}

	modules := make([]types.AuditModuleOption, 0, len(moduleSeen))
	for value, label := range moduleSeen {
		modules = append(modules, types.AuditModuleOption{Value: value, Label: label})
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Label < modules[j].Label
	})

	actions := make([]types.AuditActionOption, 0, len(actionSeen))
	for value, label := range actionSeen {
		actions = append(actions, types.AuditActionOption{Value: value, Label: label})
	}
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Label < actions[j].Label
	})
	return modules, actions, nil
}

func (s *AuditLogService) loadParentPermissions(ctx context.Context, apis []*model.Permission) (map[uint]*model.Permission, error) {
	seen := make(map[uint]struct{}, len(apis))
	parentIDs := make([]uint, 0, len(apis))
	for _, api := range apis {
		if api.ParentID == 0 {
			continue
		}
		if _, ok := seen[api.ParentID]; ok {
			continue
		}
		seen[api.ParentID] = struct{}{}
		parentIDs = append(parentIDs, api.ParentID)
	}
	if len(parentIDs) == 0 {
		return map[uint]*model.Permission{}, nil
	}
	parents, err := s.permissionRepo.GetByIDs(ctx, parentIDs)
	if err != nil {
		return nil, fmt.Errorf("调用 GetByIDs 失败: %w", err)
	}
	parentMap := make(map[uint]*model.Permission, len(parents))
	for _, p := range parents {
		parentMap[p.ID] = p
	}
	return parentMap, nil
}

func auditActionFromCode(code string) string {
	if idx := strings.LastIndex(code, ":"); idx >= 0 {
		return code[idx+1:]
	}
	return code
}

func auditModuleFromMenuCode(code string) string {
	return strings.ReplaceAll(code, ":", "_")
}

func auditModuleFromCode(code string) string {
	if idx := strings.LastIndex(code, ":"); idx > 0 {
		return strings.ReplaceAll(code[:idx], ":", "_")
	}
	return code
}

func (s *AuditLogService) Get(ctx context.Context, id uint) (*types.AuditLogDetailResponse, error) {
	row, err := s.auditLogRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("审计记录不存在")
		}
		s.log.Error(ctx, "查询操作审计详情失败", logger.Error(err))
		return nil, fmt.Errorf("查询操作审计详情失败: %w", err)
	}
	item := toAuditLogItem(row)
	return &item, nil
}

func toAuditLogItem(row *model.AuditLog) types.AuditLogItem {
	successText := "失败"
	if row.Success {
		successText = "成功"
	}
	return types.AuditLogItem{
		ID:             row.ID,
		EmployeeID:     row.EmployeeID,
		EmployeeName:   row.EmployeeName,
		EmployeeMobile: row.EmployeeMobile,
		Module:         row.Module,
		ModuleText:     row.ModuleText,
		Action:         row.Action,
		ActionText:     row.ActionText,
		TargetType:     row.TargetType,
		TargetID:       row.TargetID,
		TargetLabel:    row.TargetLabel,
		RequestPath:    row.RequestPath,
		RequestMethod:  row.RequestMethod,
		Success:        row.Success,
		SuccessText:    successText,
		Detail:         row.Detail,
		IP:             row.IP,
		UserAgent:      row.UserAgent,
		CreatedAt:      utils.FormatDateTime(row.CreatedAt),
	}
}

func parseAuditTimeRange(start, end string) (*time.Time, *time.Time, error) {
	var startTime *time.Time
	var endTime *time.Time
	if start != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", start, time.Local)
		if err != nil {
			return nil, nil, errs.Invalid("开始时间格式错误")
		}
		startTime = &t
	}
	if end != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", end, time.Local)
		if err != nil {
			return nil, nil, errs.Invalid("结束时间格式错误")
		}
		endTime = &t
	}
	return startTime, endTime, nil
}

func parseAuditTarget(body []byte) (uint, string) {
	if len(body) == 0 {
		return 0, ""
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, ""
	}

	targetID := uint(0)
	for _, key := range []string{"id", "employeeId", "roleId"} {
		if v, ok := payload[key]; ok {
			if id := toUint(v); id > 0 {
				targetID = id
				break
			}
		}
	}

	labels := make([]string, 0, 6)
	if targetID > 0 {
		labels = append(labels, "ID "+strconv.FormatUint(uint64(targetID), 10))
	}
	if v, ok := payload["mobile"]; ok {
		if s := toString(v); s != "" {
			labels = append(labels, s)
		}
	}
	if v, ok := payload["name"]; ok {
		if s := toString(v); s != "" {
			labels = append(labels, s)
		}
	}
	if v, ok := payload["approve"]; ok {
		if b, ok := v.(bool); ok {
			if b {
				labels = append(labels, "通过")
			} else {
				labels = append(labels, "驳回")
			}
		}
	}
	if v, ok := payload["status"]; ok {
		labels = append(labels, "status="+toString(v))
	}
	if v, ok := payload["duration"]; ok {
		labels = append(labels, "duration="+toString(v))
	}
	if v, ok := payload["roleIds"]; ok {
		labels = append(labels, "roles="+toString(v))
	}
	if v, ok := payload["permissionIds"]; ok {
		labels = append(labels, "permissions="+toString(v))
	}

	return targetID, strings.Join(labels, " · ")
}

var auditSensitiveKeys = map[string]struct{}{
	"mobile": {}, "password": {}, "code": {}, "refreshToken": {}, "accessToken": {},
}

func sanitizeAuditDetail(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return truncateAuditDetail(string(body))
	}
	for key := range payload {
		if _, ok := auditSensitiveKeys[key]; ok {
			payload[key] = "***"
		}
	}
	sanitized, err := json.Marshal(payload)
	if err != nil {
		return truncateAuditDetail(string(body))
	}
	return truncateAuditDetail(string(sanitized))
}

func toUint(v any) uint {
	switch n := v.(type) {
	case float64:
		if n > 0 {
			return uint(n)
		}
	case json.Number:
		if i, err := n.Int64(); err == nil && i > 0 {
			return uint(i)
		}
	case string:
		if i, err := strconv.ParseUint(n, 10, 64); err == nil && i > 0 {
			return uint(i)
		}
	}
	return 0
}

func toString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case float64:
		return strconv.FormatInt(int64(s), 10)
	case bool:
		if s {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(v)
	}
}

func truncateAuditDetail(s string) string {
	if len(s) <= maxAuditDetailLen {
		return s
	}
	return s[:maxAuditDetailLen] + "..."
}
