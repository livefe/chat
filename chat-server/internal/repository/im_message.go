package repository

import (
	"context"
	"errors"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type IMMessageRepository interface {
	Create(ctx context.Context, msg *model.IMMessage) error
	GetByID(ctx context.Context, id uint) (*model.IMMessage, error)
	GetByClientMsgID(ctx context.Context, senderKind uint8, senderID uint, clientMsgID string) (*model.IMMessage, error)
	ListByConversation(ctx context.Context, conversationID, beforeID uint, limit int) ([]*model.IMMessage, error)
	LastVisible(ctx context.Context, conversationID uint) (*model.IMMessage, error)
	MarkRecalled(ctx context.Context, id uint) error
}

type imMessageRepository struct {
	db *database.DB
}

func NewIMMessageRepository(db *database.DB) IMMessageRepository {
	return &imMessageRepository{db: db}
}

func (r *imMessageRepository) Create(ctx context.Context, msg *model.IMMessage) error {
	return database.Conn(ctx, r.db).Create(msg).Error
}

func (r *imMessageRepository) GetByID(ctx context.Context, id uint) (*model.IMMessage, error) {
	var msg model.IMMessage
	err := database.Conn(ctx, r.db).First(&msg, id).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *imMessageRepository) GetByClientMsgID(ctx context.Context, senderKind uint8, senderID uint, clientMsgID string) (*model.IMMessage, error) {
	if clientMsgID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var msg model.IMMessage
	err := database.Conn(ctx, r.db).
		Where("sender_kind = ? AND sender_id = ? AND client_msg_id = ?", senderKind, senderID, clientMsgID).
		First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *imMessageRepository) ListByConversation(ctx context.Context, conversationID, beforeID uint, limit int) ([]*model.IMMessage, error) {
	db := database.Conn(ctx, r.db).Where("conversation_id = ?", conversationID)
	if beforeID > 0 {
		db = db.Where("id < ?", beforeID)
	}
	var list []*model.IMMessage
	err := db.Order("id DESC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *imMessageRepository) LastVisible(ctx context.Context, conversationID uint) (*model.IMMessage, error) {
	var msg model.IMMessage
	err := database.Conn(ctx, r.db).
		Where("conversation_id = ? AND recalled = 0", conversationID).
		Order("id DESC").
		First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *imMessageRepository) MarkRecalled(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Model(&model.IMMessage{}).
		Where("id = ?", id).
		Update("recalled", 1).Error
}
