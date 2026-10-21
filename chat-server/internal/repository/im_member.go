package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type IMConversationRow struct {
	model.IMConversation
	Pinned        bool `gorm:"column:pinned"`
	PeerID        uint `gorm:"column:peer_id"`
	LastReadMsgID uint `gorm:"column:last_read_msg_id"`
}

type IMMemberRepository interface {
	CreateBatch(ctx context.Context, members []*model.IMMember) error
	GetActive(ctx context.Context, conversationID uint, kind uint8, principalID uint) (*model.IMMember, error)
	ListActive(ctx context.Context, conversationID uint) ([]*model.IMMember, error)
	UpdateRead(ctx context.Context, conversationID uint, kind uint8, principalID, lastReadMsgID uint) error
	Leave(ctx context.Context, conversationID uint, kind uint8, principalID uint, leftAt int64) error
	Reactivate(ctx context.Context, conversationID uint, kind uint8, principalID uint, role uint8, joinedAt int64) error
	ListConversationsByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]IMConversationRow, int64, error)
}

type imMemberRepository struct {
	db *database.DB
}

func NewIMMemberRepository(db *database.DB) IMMemberRepository {
	return &imMemberRepository{db: db}
}

func (r *imMemberRepository) CreateBatch(ctx context.Context, members []*model.IMMember) error {
	if len(members) == 0 {
		return nil
	}
	return database.Conn(ctx, r.db).Create(&members).Error
}

func (r *imMemberRepository) GetActive(ctx context.Context, conversationID uint, kind uint8, principalID uint) (*model.IMMember, error) {
	var m model.IMMember
	err := database.Conn(ctx, r.db).
		Where("conversation_id = ? AND principal_kind = ? AND principal_id = ? AND left_at = 0", conversationID, kind, principalID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *imMemberRepository) ListActive(ctx context.Context, conversationID uint) ([]*model.IMMember, error) {
	var list []*model.IMMember
	err := database.Conn(ctx, r.db).
		Where("conversation_id = ? AND left_at = 0", conversationID).
		Find(&list).Error
	return list, err
}

func (r *imMemberRepository) UpdateRead(ctx context.Context, conversationID uint, kind uint8, principalID, lastReadMsgID uint) error {
	return database.Conn(ctx, r.db).Model(&model.IMMember{}).
		Where("conversation_id = ? AND principal_kind = ? AND principal_id = ? AND left_at = 0", conversationID, kind, principalID).
		Update("last_read_msg_id", lastReadMsgID).Error
}

func (r *imMemberRepository) Leave(ctx context.Context, conversationID uint, kind uint8, principalID uint, leftAt int64) error {
	return database.Conn(ctx, r.db).Model(&model.IMMember{}).
		Where("conversation_id = ? AND principal_kind = ? AND principal_id = ? AND left_at = 0", conversationID, kind, principalID).
		Update("left_at", leftAt).Error
}

func (r *imMemberRepository) Reactivate(ctx context.Context, conversationID uint, kind uint8, principalID uint, role uint8, joinedAt int64) error {
	res := database.Conn(ctx, r.db).Model(&model.IMMember{}).
		Where("conversation_id = ? AND principal_kind = ? AND principal_id = ?", conversationID, kind, principalID).
		Updates(map[string]any{
			"left_at":   0,
			"role":      role,
			"joined_at": joinedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	return r.CreateBatch(ctx, []*model.IMMember{{
		ConversationID: conversationID,
		PrincipalKind:  kind,
		PrincipalID:    principalID,
		Role:           role,
		JoinedAt:       joinedAt,
	}})
}

func (r *imMemberRepository) ListConversationsByEmployee(ctx context.Context, employeeID uint, page, pageSize int) ([]IMConversationRow, int64, error) {
	base := database.Conn(ctx, r.db).Table("im_conversation AS c").
		Joins(`INNER JOIN im_member AS m ON m.conversation_id = c.id AND m.principal_kind = ? AND m.principal_id = ? AND m.left_at = 0 AND m.deleted_at IS NULL`,
			model.IMPrincipalEmployee, employeeID).
		Joins(`LEFT JOIN im_member AS peer ON peer.conversation_id = c.id AND peer.principal_kind = ? AND peer.principal_id <> ? AND peer.left_at = 0 AND peer.deleted_at IS NULL`,
			model.IMPrincipalEmployee, employeeID).
		Where("c.deleted_at IS NULL").
		Where("c.type = ?", model.IMConvTypeStaffDM)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []IMConversationRow
	err := base.Select(`c.*, CASE WHEN m.pinned = 1 THEN 1 ELSE 0 END AS pinned, COALESCE(peer.principal_id, 0) AS peer_id, m.last_read_msg_id AS last_read_msg_id`).
		Order("pinned DESC, c.last_active_at DESC, c.id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error
	return list, total, err
}
