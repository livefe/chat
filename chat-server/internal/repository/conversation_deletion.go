package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type ConversationDeletionRepository interface {
	Upsert(ctx context.Context, row *model.ConversationDeletion) error
	Get(ctx context.Context, userID, peerID uint) (*model.ConversationDeletion, error)
	DeleteByUserID(ctx context.Context, userID uint) error
}

type conversationDeletionRepository struct {
	db *database.DB
}

func NewConversationDeletionRepository(db *database.DB) ConversationDeletionRepository {
	return &conversationDeletionRepository{db: db}
}

func (r *conversationDeletionRepository) Upsert(ctx context.Context, row *model.ConversationDeletion) error {
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "peer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"cleared_before_id", "updated_at"}),
	}).Create(row).Error
}

func (r *conversationDeletionRepository) Get(ctx context.Context, userID, peerID uint) (*model.ConversationDeletion, error) {
	var row model.ConversationDeletion
	err := database.Conn(ctx, r.db).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *conversationDeletionRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return database.Conn(ctx, r.db).
		Where("user_id = ? OR peer_id = ?", userID, userID).
		Delete(&model.ConversationDeletion{}).Error
}
