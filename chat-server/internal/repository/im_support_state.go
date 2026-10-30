package repository

import (
	"context"
	"strconv"
	"strings"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

const supportListColumns = "c.*, s.user_id, s.assignee_employee_id, s.queue_at, s.accept_at, s.close_at, s.close_reason, s.close_note"

type SupportListRow struct {
	model.IMConversation
	UserID             uint   `gorm:"column:user_id"`
	AssigneeEmployeeID uint   `gorm:"column:assignee_employee_id"`
	QueueAt            int64  `gorm:"column:queue_at"`
	AcceptAt           int64  `gorm:"column:accept_at"`
	CloseAt            int64  `gorm:"column:close_at"`
	CloseReason        string `gorm:"column:close_reason"`
	CloseNote          string `gorm:"column:close_note"`
}

type SupportOverviewFilter struct {
	AssigneeEmployeeID uint
	Keyword            string
	From               int64
	To                 int64
}

type IMSupportStateRepository interface {
	Create(ctx context.Context, st *model.IMSupportState) error
	GetByConversationID(ctx context.Context, conversationID uint) (*model.IMSupportState, error)
	GetOpenByUserID(ctx context.Context, userID uint) (*model.IMSupportState, error)
	GetLatestByUserID(ctx context.Context, userID uint) (*model.IMSupportState, error)
	Update(ctx context.Context, conversationID uint, updates map[string]any) error
	GetRow(ctx context.Context, conversationID uint) (*SupportListRow, error)
	ListQueue(ctx context.Context, page, pageSize int, keyword string) ([]SupportListRow, int64, error)
	CountQueue(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	CountQueueLongWait(ctx context.Context, waitSec int64, nowUnix int64) (int64, error)
	ListByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error)
	ListClosedByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error)
	ListActive(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error)
	ListClosed(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error)
	ListByUser(ctx context.Context, userID, excludeConversationID uint, page, pageSize int) ([]SupportListRow, int64, error)
	HasServedUser(ctx context.Context, userID, employeeID uint) (bool, error)
	ListAgentEmployeeIDs(ctx context.Context, permissionCode, superRoleCode string) ([]uint, error)
}

type imSupportStateRepository struct {
	db *database.DB
}

func NewIMSupportStateRepository(db *database.DB) IMSupportStateRepository {
	return &imSupportStateRepository{db: db}
}

func (r *imSupportStateRepository) Create(ctx context.Context, st *model.IMSupportState) error {
	return database.Conn(ctx, r.db).Create(st).Error
}

func (r *imSupportStateRepository) GetByConversationID(ctx context.Context, conversationID uint) (*model.IMSupportState, error) {
	var st model.IMSupportState
	err := database.Conn(ctx, r.db).Where("conversation_id = ?", conversationID).First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *imSupportStateRepository) GetOpenByUserID(ctx context.Context, userID uint) (*model.IMSupportState, error) {
	var st model.IMSupportState
	err := database.Conn(ctx, r.db).
		Joins("INNER JOIN im_conversation AS c ON c.id = im_support_state.conversation_id AND c.deleted_at IS NULL").
		Where("im_support_state.user_id = ? AND c.type = ? AND c.status IN ?", userID, model.IMConvTypeSupport, []int{int(model.IMConvStatusQueued), int(model.IMConvStatusActive)}).
		Order("im_support_state.id DESC").
		First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *imSupportStateRepository) GetLatestByUserID(ctx context.Context, userID uint) (*model.IMSupportState, error) {
	var st model.IMSupportState
	err := database.Conn(ctx, r.db).
		Joins("INNER JOIN im_conversation AS c ON c.id = im_support_state.conversation_id AND c.deleted_at IS NULL").
		Where("im_support_state.user_id = ? AND c.type = ?", userID, model.IMConvTypeSupport).
		Order("im_support_state.id DESC").
		First(&st).Error
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *imSupportStateRepository) Update(ctx context.Context, conversationID uint, updates map[string]any) error {
	return database.Conn(ctx, r.db).Model(&model.IMSupportState{}).
		Where("conversation_id = ?", conversationID).
		Updates(updates).Error
}

func (r *imSupportStateRepository) supportBase(ctx context.Context) *gorm.DB {
	return database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ?", model.IMConvTypeSupport)
}

func pageRows(base *gorm.DB, order string, page, pageSize int) ([]SupportListRow, int64, error) {
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SupportListRow
	err := base.Select(supportListColumns).
		Order(order).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func withSupportKeyword(base *gorm.DB, keyword string) *gorm.DB {
	kw := strings.TrimSpace(keyword)
	if kw == "" {
		return base
	}
	base = base.Joins("LEFT JOIN `user` AS u ON u.id = s.user_id AND u.deleted_at IS NULL")
	like := "%" + kw + "%"
	if id, err := strconv.ParseUint(kw, 10, 64); err == nil {
		return base.Where("(s.user_id = ? OR c.id = ? OR u.mobile LIKE ? OR u.nickname LIKE ?)", id, id, like, like)
	}
	return base.Where("(u.mobile LIKE ? OR u.nickname LIKE ?)", like, like)
}

func (r *imSupportStateRepository) GetRow(ctx context.Context, conversationID uint) (*SupportListRow, error) {
	var row SupportListRow
	res := r.supportBase(ctx).Where("c.id = ?", conversationID).Select(supportListColumns).Limit(1).Scan(&row)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *imSupportStateRepository) ListQueue(ctx context.Context, page, pageSize int, keyword string) ([]SupportListRow, int64, error) {
	base := withSupportKeyword(r.supportBase(ctx).Where("c.status = ?", model.IMConvStatusQueued), keyword)
	return pageRows(base, "s.queue_at ASC, c.id ASC", page, pageSize)
}

func (r *imSupportStateRepository) CountQueue(ctx context.Context) (int64, error) {
	var total int64
	err := r.supportBase(ctx).Where("c.status = ?", model.IMConvStatusQueued).Count(&total).Error
	return total, err
}

func (r *imSupportStateRepository) CountActive(ctx context.Context) (int64, error) {
	var total int64
	err := r.supportBase(ctx).Where("c.status = ?", model.IMConvStatusActive).Count(&total).Error
	return total, err
}

func (r *imSupportStateRepository) CountQueueLongWait(ctx context.Context, waitSec int64, nowUnix int64) (int64, error) {
	if waitSec < 0 {
		waitSec = 0
	}
	var total int64
	err := r.supportBase(ctx).
		Where("c.status = ? AND s.queue_at > 0 AND s.queue_at <= ?", model.IMConvStatusQueued, nowUnix-waitSec).
		Count(&total).Error
	return total, err
}

func (r *imSupportStateRepository) ListByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error) {
	base := r.supportBase(ctx).Where("c.status = ? AND s.assignee_employee_id = ?", model.IMConvStatusActive, employeeID)
	return pageRows(base, "c.last_active_at DESC, c.id DESC", page, pageSize)
}

func (r *imSupportStateRepository) ListClosedByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error) {
	base := r.supportBase(ctx).Where("c.status = ? AND s.assignee_employee_id = ?", model.IMConvStatusClosed, employeeID)
	return pageRows(base, "s.close_at DESC, c.id DESC", page, pageSize)
}

func (r *imSupportStateRepository) overviewBase(ctx context.Context, status uint8, timeColumn string, filter SupportOverviewFilter) *gorm.DB {
	base := r.supportBase(ctx).Where("c.status = ?", status)
	if filter.AssigneeEmployeeID > 0 {
		base = base.Where("s.assignee_employee_id = ?", filter.AssigneeEmployeeID)
	}
	if filter.From > 0 {
		base = base.Where(timeColumn+" >= ?", filter.From)
	}
	if filter.To > 0 {
		base = base.Where(timeColumn+" < ?", filter.To)
	}
	return withSupportKeyword(base, filter.Keyword)
}

func (r *imSupportStateRepository) ListActive(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error) {
	base := r.overviewBase(ctx, model.IMConvStatusActive, "s.accept_at", filter)
	return pageRows(base, "c.last_active_at DESC, c.id DESC", page, pageSize)
}

func (r *imSupportStateRepository) ListClosed(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error) {
	base := r.overviewBase(ctx, model.IMConvStatusClosed, "s.close_at", filter)
	return pageRows(base, "s.close_at DESC, c.id DESC", page, pageSize)
}

func (r *imSupportStateRepository) ListByUser(ctx context.Context, userID, excludeConversationID uint, page, pageSize int) ([]SupportListRow, int64, error) {
	base := r.supportBase(ctx).Where("s.user_id = ?", userID)
	if excludeConversationID > 0 {
		base = base.Where("c.id <> ?", excludeConversationID)
	}
	return pageRows(base, "c.id DESC", page, pageSize)
}

func (r *imSupportStateRepository) HasServedUser(ctx context.Context, userID, employeeID uint) (bool, error) {
	var n int64
	err := r.supportBase(ctx).
		Where("s.user_id = ? AND s.assignee_employee_id = ?", userID, employeeID).
		Count(&n).Error
	return n > 0, err
}

func (r *imSupportStateRepository) ListAgentEmployeeIDs(ctx context.Context, permissionCode, superRoleCode string) ([]uint, error) {
	var ids []uint
	err := database.Conn(ctx, r.db).Table("employee AS e").
		Joins("INNER JOIN employee_role AS er ON er.employee_id = e.id AND er.deleted_at IS NULL").
		Joins("INNER JOIN role AS ro ON ro.id = er.role_id AND ro.deleted_at IS NULL AND ro.status = ?", model.RoleStatusEnabled).
		Where("e.deleted_at IS NULL AND e.status = ?", model.EmployeeStatusEnabled).
		Where(`(ro.code = ? OR EXISTS (
			SELECT 1 FROM role_permission AS rp
			INNER JOIN permission AS p ON p.id = rp.permission_id AND p.deleted_at IS NULL AND p.status = ?
			WHERE rp.role_id = ro.id AND rp.deleted_at IS NULL AND p.code = ?))`,
			superRoleCode, model.PermissionStatusEnabled, permissionCode).
		Distinct("e.id").
		Order("e.id ASC").
		Pluck("e.id", &ids).Error
	return ids, err
}
