package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type ConversationListFilter struct {
	UserID   uint
	Page     int
	PageSize int
}

type ConversationListRow struct {
	model.Conversation
	Pinned bool `gorm:"column:pinned"`
}

type ConversationRepository interface {
	Upsert(ctx context.Context, conv *model.Conversation) error
	List(ctx context.Context, f ConversationListFilter) ([]*model.Conversation, int64, error)
	ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]ConversationListRow, int64, error)
	UpdateLastPreview(ctx context.Context, userMinID, userMaxID, lastID uint, lastContent string) error
}

type conversationRepository struct {
	db *database.DB
}

func NewConversationRepository(db *database.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Upsert(ctx context.Context, conv *model.Conversation) error {
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_min_id"}, {Name: "user_max_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"last_message_id", "last_content", "last_active_at", "updated_at",
		}),
	}).Create(conv).Error
}

func (r *conversationRepository) List(ctx context.Context, f ConversationListFilter) ([]*model.Conversation, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.Conversation{})
	if f.UserID > 0 {
		db = db.Where("user_min_id = ? OR user_max_id = ?", f.UserID, f.UserID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.Conversation
	if err := db.Order("last_active_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *conversationRepository) ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]ConversationListRow, int64, error) {
	base := database.Conn(ctx, r.db).Model(&model.Conversation{}).
		Joins(
			`LEFT JOIN conversation_deletion AS cd ON cd.user_id = ? AND cd.peer_id = CASE WHEN conversation.user_min_id = ? THEN conversation.user_max_id ELSE conversation.user_min_id END`,
			userID, userID,
		).
		Joins(
			`LEFT JOIN conversation_pin AS cp ON cp.user_id = ? AND cp.peer_id = CASE WHEN conversation.user_min_id = ? THEN conversation.user_max_id ELSE conversation.user_min_id END`,
			userID, userID,
		).
		Where("(conversation.user_min_id = ? OR conversation.user_max_id = ?)", userID, userID).
		Where("cd.id IS NULL OR conversation.last_message_id > cd.cleared_before_id")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []ConversationListRow
	if err := base.Select("conversation.*, CASE WHEN cp.id IS NOT NULL THEN 1 ELSE 0 END AS pinned").
		Order("pinned DESC, conversation.last_active_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *conversationRepository) UpdateLastPreview(ctx context.Context, userMinID, userMaxID, lastID uint, lastContent string) error {
	return database.Conn(ctx, r.db).Model(&model.Conversation{}).
		Where("user_min_id = ? AND user_max_id = ?", userMinID, userMaxID).
		Updates(map[string]any{
			"last_message_id": lastID,
			"last_content":    lastContent,
		}).Error
}
