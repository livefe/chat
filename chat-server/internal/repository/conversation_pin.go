package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type ConversationPinRepository interface {
	Upsert(ctx context.Context, row *model.ConversationPin) error
	Delete(ctx context.Context, userID, peerID uint) error
	Exists(ctx context.Context, userID, peerID uint) (bool, error)
	DeleteByUserID(ctx context.Context, userID uint) error
}

type conversationPinRepository struct {
	db *database.DB
}

func NewConversationPinRepository(db *database.DB) ConversationPinRepository {
	return &conversationPinRepository{db: db}
}

func (r *conversationPinRepository) Upsert(ctx context.Context, row *model.ConversationPin) error {
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "peer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).Create(row).Error
}

func (r *conversationPinRepository) Delete(ctx context.Context, userID, peerID uint) error {
	return database.Conn(ctx, r.db).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		Delete(&model.ConversationPin{}).Error
}

func (r *conversationPinRepository) Exists(ctx context.Context, userID, peerID uint) (bool, error) {
	var count int64
	err := database.Conn(ctx, r.db).Model(&model.ConversationPin{}).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *conversationPinRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return database.Conn(ctx, r.db).
		Where("user_id = ?", userID).
		Delete(&model.ConversationPin{}).Error
}
