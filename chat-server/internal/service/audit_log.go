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
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"gorm.io/gorm"
)

const (
	maxAuditDetailLen        = 2000
	maxAuditResultMessageLen = 255
	auditEmployeeOptionLimit = 20
)

type AuditLogInput struct {
	EmployeeID    uint
	RequestPath   string
	RequestMethod string
	RequestBody   []byte
	IP            string
	UserAgent     string
	Success       bool
	ResultCode    int
	ResultMessage string
}

var auditTargetTypeByModule = map[string]string{
	"user:account":        "user",
	"user:feedback":       "feedback",
	"user:report":         "report",
	"org:employee":        "employee",
	"org:department":      "department",
	"org:role":            "role",
	"org:permission":      "permission",
	"ops":                 "media",
	"ops:region":          "region",
	"ops:app_release":     "app_release",
	"ops:sensitive_word":  "sensitive_word",
	"im:chat":             "im_conversation",
	"im:support":          "support_session",
	"im:support_overview": "support_config",
	"chat:room":           "room",
	"chat:message":        "message",
	"chat:conversation":   "conversation",
	"post:post":           "post",
	"ad:campaign":         "ad_campaign",
	"ad:setting":          "ad_setting",
}

var auditTargetTypeText = map[string]string{
	"user":            "用户",
	"feedback":        "反馈",
	"report":          "举报",
	"employee":        "员工",
	"department":      "部门",
	"role":            "角色",
	"permission":      "权限",
	"media":           "媒体文件",
	"region":          "行政区划",
	"app_release":     "应用版本",
	"sensitive_word":  "屏蔽词",
	"im_conversation": "员工会话",
	"support_session": "客服会话",
	"support_config":  "客服配置",
	"room":            "聊天区",
	"message":         "聊天消息",
	"conversation":    "私信会话",
	"post":            "帖子",
	"ad_campaign":     "广告计划",
	"ad_setting":      "投放设置",
}

var auditTargetKeys = []struct {
	key        string
	targetType string
}{
	{"id", ""},
	{"ids", ""},
	{"employeeId", "employee"},
	{"roleId", "role"},
	{"userId", "user"},
	{"postId", "post"},
	{"roomId", "room"},
	{"reportId", "report"},
	{"feedbackId", "feedback"},
	{"campaignId", "ad_campaign"},
	{"messageId", "message"},
	{"conversationId", ""},
}

type AuditLogService struct {
	log            *logger.Logger
	auditLogRepo   repository.AuditLogRepository
	employeeRepo   repository.EmployeeRepository
	permissionRepo repository.PermissionRepository

	excluded map[string]struct{}
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
		excluded:       map[string]struct{}{},
	}
}

func auditRouteKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func (s *AuditLogService) ExcludeRoutes(routes ...string) {
	for _, r := range routes {
		method, path, ok := strings.Cut(strings.TrimSpace(r), " ")
		if !ok {
			continue
		}
		s.excluded[auditRouteKey(method, strings.TrimSpace(path))] = struct{}{}
	}
}

func (s *AuditLogService) Excluded(method, path string) bool {
	_, ok := s.excluded[auditRouteKey(method, path)]
	return ok
}

type auditMeta struct {
	module     string
	moduleText string
	action     string
	actionText string
	targetType string
}

func (s *AuditLogService) Record(ctx context.Context, in AuditLogInput) {
	if s.Excluded(in.RequestMethod, in.RequestPath) {
		return
	}
	meta, ok := s.resolveAuditMeta(ctx, in.RequestPath, in.RequestMethod)
	if !ok {
		return
	}

	employee, err := s.employeeRepo.GetByID(ctx, in.EmployeeID)
	if err != nil {
		s.log.Error(ctx, "写入操作审计失败：查询员工失败", logger.Uint("employee_id", in.EmployeeID), logger.Error(err))
		return
	}

	target := parseAuditTarget(in.RequestBody)
	targetType := meta.targetType
	if target.targetType != "" {
		targetType = target.targetType
	}

	row := &model.AuditLog{
		EmployeeID:     in.EmployeeID,
		EmployeeName:   employee.Name,
		EmployeeMobile: employee.Mobile,
		Module:         meta.module,
		ModuleText:     meta.moduleText,
		Action:         meta.action,
		ActionText:     meta.actionText,
		TargetType:     targetType,
		TargetID:       target.id,
		TargetLabel:    clipRunes(target.label, 128),
		RequestPath:    in.RequestPath,
		RequestMethod:  in.RequestMethod,
		Success:        in.Success,
		Detail:         sanitizeAuditDetail(in.RequestBody),
		IP:             in.IP,
		UserAgent:      clipRunes(in.UserAgent, 500),
	}
	if !in.Success {
		row.ResultCode = in.ResultCode
		row.ResultMessage = clipRunes(in.ResultMessage, maxAuditResultMessageLen)
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
		EmployeeID:  req.EmployeeID,
		Module:      strings.TrimSpace(req.Module),
		Action:      strings.TrimSpace(req.Action),
		TargetType:  strings.TrimSpace(req.TargetType),
		TargetID:    req.TargetID,
		RequestPath: strings.TrimSpace(req.RequestPath),
		Success:     req.Success,
		Keyword:     strings.TrimSpace(req.Keyword),
		StartTime:   startTime,
		EndTime:     endTime,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询操作审计失败", logger.Error(err))
		return nil, fmt.Errorf("查询操作审计失败: %w", err)
	}

	list := make([]types.AuditLogItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, toAuditLogItem(row))
	}
	return &types.AuditLogListResponse{Total: total, List: list}, nil
}

func (s *AuditLogService) Options(ctx context.Context) (*types.AuditOptionsResponse, error) {
	apis, err := s.permissionRepo.ListEnabledAPIs(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询审计筛选项失败: %w", err)
	}
	parentMap, err := s.loadParentPermissions(ctx, apis)
	if err != nil {
		return nil, fmt.Errorf("查询审计筛选项失败: %w", err)
	}

	moduleSeen := make(map[string]string)
	actions := make([]types.AuditActionOption, 0, len(apis))
	actionSeen := make(map[string]struct{})
	for _, api := range apis {
		if strings.ToUpper(strings.TrimSpace(api.Method)) != http.MethodPost {
			continue
		}
		if s.Excluded(api.Method, api.Path) {
			continue
		}
		meta := auditMetaFromPermission(api, parentMap[api.ParentID])
		moduleSeen[meta.module] = meta.moduleText
		if _, ok := actionSeen[meta.action]; ok {
			continue
		}
		actionSeen[meta.action] = struct{}{}
		actions = append(actions, types.AuditActionOption{Value: meta.action, Label: meta.actionText, Module: meta.module})
	}

	modules := make([]types.AuditOption, 0, len(moduleSeen))
	for value, label := range moduleSeen {
		modules = append(modules, types.AuditOption{Value: value, Label: label})
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].Value < modules[j].Value })
	sort.Slice(actions, func(i, j int) bool { return actions[i].Value < actions[j].Value })

	targetTypes := make([]types.AuditOption, 0, len(auditTargetTypeText))
	for value, label := range auditTargetTypeText {
		targetTypes = append(targetTypes, types.AuditOption{Value: value, Label: label})
	}
	sort.Slice(targetTypes, func(i, j int) bool { return targetTypes[i].Value < targetTypes[j].Value })

	return &types.AuditOptionsResponse{Modules: modules, Actions: actions, TargetTypes: targetTypes}, nil
}

func (s *AuditLogService) EmployeeOptions(ctx context.Context, req *types.AuditEmployeeOptionsRequest) ([]types.AuditEmployeeOption, error) {
	var employees []*model.Employee
	if req.ID > 0 {
		e, err := s.employeeRepo.GetByID(ctx, req.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询员工失败: %w", err)
		}
		if e != nil {
			employees = append(employees, e)
		}
	} else {
		list, _, err := s.employeeRepo.List(ctx, repository.EmployeeListFilter{
			Keyword:  strings.TrimSpace(req.Keyword),
			Page:     1,
			PageSize: auditEmployeeOptionLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("查询员工失败: %w", err)
		}
		employees = list
	}
	out := make([]types.AuditEmployeeOption, 0, len(employees))
	for _, e := range employees {
		out = append(out, types.AuditEmployeeOption{ID: e.ID, Name: e.Name, Mobile: e.Mobile})
	}
	return out, nil
}

func (s *AuditLogService) resolveAuditMeta(ctx context.Context, path, method string) (auditMeta, bool) {
	perm, err := s.permissionRepo.GetAPIByPathAndMethod(ctx, path, method)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error(ctx, "查询审计元数据失败", logger.Error(err))
		}
		return auditMeta{}, false
	}
	var parent *model.Permission
	if perm.ParentID > 0 {
		parent, err = s.permissionRepo.GetByID(ctx, perm.ParentID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Error(ctx, "查询审计父级权限失败", logger.Uint("parent_id", perm.ParentID), logger.Error(err))
		}
	}
	return auditMetaFromPermission(perm, parent), true
}

func auditMetaFromPermission(perm, parent *model.Permission) auditMeta {
	m := auditMeta{
		module:     perm.Code,
		moduleText: perm.Name,
		action:     perm.Code,
		actionText: perm.Name,
	}
	if parent != nil {
		m.module = parent.Code
		m.moduleText = parent.Name
	}
	m.targetType = auditTargetTypeByModule[m.module]
	return m
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
	targetTypeText := auditTargetTypeText[row.TargetType]
	if targetTypeText == "" {
		targetTypeText = row.TargetType
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
		TargetTypeText: targetTypeText,
		TargetID:       row.TargetID,
		TargetLabel:    row.TargetLabel,
		RequestPath:    row.RequestPath,
		RequestMethod:  row.RequestMethod,
		Success:        row.Success,
		SuccessText:    successText,
		ResultCode:     row.ResultCode,
		ResultMessage:  row.ResultMessage,
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

type auditTarget struct {
	id         uint
	targetType string
	label      string
}

func parseAuditTarget(body []byte) auditTarget {
	var out auditTarget
	if len(body) == 0 {
		return out
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return out
	}

	labels := make([]string, 0, 6)
	for _, k := range auditTargetKeys {
		v, ok := payload[k.key]
		if !ok {
			continue
		}
		if list, isList := v.([]any); isList {
			ids := make([]uint, 0, len(list))
			for _, item := range list {
				if id := toUint(item); id > 0 {
					ids = append(ids, id)
				}
			}
			if len(ids) == 0 {
				continue
			}
			out.id = ids[0]
			out.targetType = k.targetType
			labels = append(labels, fmt.Sprintf("共 %d 项", len(ids)))
			break
		}
		if id := toUint(v); id > 0 {
			out.id = id
			out.targetType = k.targetType
			labels = append(labels, "ID "+strconv.FormatUint(uint64(id), 10))
			break
		}
	}

	for _, key := range []string{"mobile", "name", "word", "title"} {
		if v, ok := payload[key]; ok {
			if s := toString(v); s != "" {
				labels = append(labels, s)
			}
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
		if list, isList := v.([]any); isList {
			labels = append(labels, fmt.Sprintf("permissions=%d 项", len(list)))
		}
	}

	out.label = strings.Join(labels, " · ")
	return out
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
	return clipRunes(s, maxAuditDetailLen) + "..."
}

func clipRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max])
}
