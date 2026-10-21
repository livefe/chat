package repository

import (
	"context"
	"strconv"
	"strings"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type SupportListRow struct {
	model.IMConversation
	UserID             uint  `gorm:"column:user_id"`
	AssigneeEmployeeID uint  `gorm:"column:assignee_employee_id"`
	QueueAt            int64 `gorm:"column:queue_at"`
	AcceptAt           int64 `gorm:"column:accept_at"`
	CloseAt            int64 `gorm:"column:close_at"`
}

type SupportOverviewFilter struct {
	AssigneeEmployeeID uint
	Keyword            string
}

type IMSupportStateRepository interface {
	Create(ctx context.Context, st *model.IMSupportState) error
	GetByConversationID(ctx context.Context, conversationID uint) (*model.IMSupportState, error)
	GetOpenByUserID(ctx context.Context, userID uint) (*model.IMSupportState, error)
	GetLatestByUserID(ctx context.Context, userID uint) (*model.IMSupportState, error)
	Update(ctx context.Context, conversationID uint, updates map[string]any) error
	ListQueue(ctx context.Context, page, pageSize int) ([]SupportListRow, int64, error)
	CountQueue(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
	CountQueueLongWait(ctx context.Context, waitSec int64, nowUnix int64) (int64, error)
	ListByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error)
	ListClosedByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error)
	ListActive(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error)
	ListClosed(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error)
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

func (r *imSupportStateRepository) ListQueue(ctx context.Context, page, pageSize int) ([]SupportListRow, int64, error) {
	base := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ?", model.IMConvTypeSupport, model.IMConvStatusQueued)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SupportListRow
	err := base.Select("c.*, s.user_id, s.assignee_employee_id, s.queue_at, s.accept_at, s.close_at").
		Order("s.queue_at ASC, c.id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func (r *imSupportStateRepository) CountQueue(ctx context.Context) (int64, error) {
	var total int64
	err := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ?", model.IMConvTypeSupport, model.IMConvStatusQueued).
		Count(&total).Error
	return total, err
}

func (r *imSupportStateRepository) CountActive(ctx context.Context) (int64, error) {
	var total int64
	err := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ?", model.IMConvTypeSupport, model.IMConvStatusActive).
		Count(&total).Error
	return total, err
}

func (r *imSupportStateRepository) CountQueueLongWait(ctx context.Context, waitSec int64, nowUnix int64) (int64, error) {
	if waitSec < 0 {
		waitSec = 0
	}
	cutoff := nowUnix - waitSec
	var total int64
	err := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ? AND s.queue_at > 0 AND s.queue_at <= ?",
			model.IMConvTypeSupport, model.IMConvStatusQueued, cutoff).
		Count(&total).Error
	return total, err
}

func (r *imSupportStateRepository) supportOverviewBase(ctx context.Context, status uint8, filter SupportOverviewFilter) *gorm.DB {
	base := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ?", model.IMConvTypeSupport, status)
	if filter.AssigneeEmployeeID > 0 {
		base = base.Where("s.assignee_employee_id = ?", filter.AssigneeEmployeeID)
	}
	kw := strings.TrimSpace(filter.Keyword)
	if kw != "" {
		base = base.Joins("LEFT JOIN `user` AS u ON u.id = s.user_id AND u.deleted_at IS NULL")
		like := "%" + kw + "%"
		if id, err := strconv.ParseUint(kw, 10, 64); err == nil {
			base = base.Where("(s.user_id = ? OR c.id = ? OR u.mobile LIKE ? OR u.nickname LIKE ?)", id, id, like, like)
		} else {
			base = base.Where("(u.mobile LIKE ? OR u.nickname LIKE ?)", like, like)
		}
	}
	return base
}

func (r *imSupportStateRepository) ListByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error) {
	base := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ? AND s.assignee_employee_id = ?",
			model.IMConvTypeSupport, model.IMConvStatusActive, employeeID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SupportListRow
	err := base.Select("c.*, s.user_id, s.assignee_employee_id, s.queue_at, s.accept_at, s.close_at").
		Order("c.last_active_at DESC, c.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func (r *imSupportStateRepository) ListClosedByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]SupportListRow, int64, error) {
	base := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins("INNER JOIN im_support_state AS s ON s.conversation_id = c.id AND s.deleted_at IS NULL").
		Where("c.deleted_at IS NULL AND c.type = ? AND c.status = ? AND s.assignee_employee_id = ?",
			model.IMConvTypeSupport, model.IMConvStatusClosed, employeeID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SupportListRow
	err := base.Select("c.*, s.user_id, s.assignee_employee_id, s.queue_at, s.accept_at, s.close_at").
		Order("s.close_at DESC, c.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func (r *imSupportStateRepository) ListActive(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error) {
	base := r.supportOverviewBase(ctx, model.IMConvStatusActive, filter)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SupportListRow
	err := base.Select("c.*, s.user_id, s.assignee_employee_id, s.queue_at, s.accept_at, s.close_at").
		Order("c.last_active_at DESC, c.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}

func (r *imSupportStateRepository) ListClosed(ctx context.Context, page, pageSize int, filter SupportOverviewFilter) ([]SupportListRow, int64, error) {
	base := r.supportOverviewBase(ctx, model.IMConvStatusClosed, filter)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SupportListRow
	err := base.Select("c.*, s.user_id, s.assignee_employee_id, s.queue_at, s.accept_at, s.close_at").
		Order("s.close_at DESC, c.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}
