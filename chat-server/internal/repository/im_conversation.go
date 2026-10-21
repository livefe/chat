package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type IMConversationRepository interface {
	Create(ctx context.Context, conv *model.IMConversation) error
	GetByID(ctx context.Context, id uint) (*model.IMConversation, error)
	UpdatePreview(ctx context.Context, id, lastMsgID uint, lastContent string, lastActiveAt int64) error
	UpdateStatus(ctx context.Context, id uint, status uint8) error
	UpdateStatusIf(ctx context.Context, id uint, from, to uint8) (bool, error)
}

type imConversationRepository struct {
	db *database.DB
}

func NewIMConversationRepository(db *database.DB) IMConversationRepository {
	return &imConversationRepository{db: db}
}

func (r *imConversationRepository) Create(ctx context.Context, conv *model.IMConversation) error {
	return database.Conn(ctx, r.db).Create(conv).Error
}

func (r *imConversationRepository) GetByID(ctx context.Context, id uint) (*model.IMConversation, error) {
	var conv model.IMConversation
	err := database.Conn(ctx, r.db).First(&conv, id).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *imConversationRepository) UpdatePreview(ctx context.Context, id, lastMsgID uint, lastContent string, lastActiveAt int64) error {
	return database.Conn(ctx, r.db).Model(&model.IMConversation{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_message_id": lastMsgID,
			"last_content":    lastContent,
			"last_active_at":  lastActiveAt,
		}).Error
}

func (r *imConversationRepository) UpdateStatus(ctx context.Context, id uint, status uint8) error {
	return database.Conn(ctx, r.db).Model(&model.IMConversation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *imConversationRepository) UpdateStatusIf(ctx context.Context, id uint, from, to uint8) (bool, error) {
	res := database.Conn(ctx, r.db).Model(&model.IMConversation{}).
		Where("id = ? AND status = ?", id, from).
		Update("status", to)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
