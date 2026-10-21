package repository

import (
	"context"
	"errors"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type PostExtraRepository interface {
	UpsertByPostID(ctx context.Context, extras *model.PostExtra) error
	DeleteByPostID(ctx context.Context, postID uint) error
	GetByPostID(ctx context.Context, postID uint) (*model.PostExtra, error)
	ListByPostIDs(ctx context.Context, postIDs []uint) ([]*model.PostExtra, error)
}

type postExtraRepository struct {
	db *database.DB
}

func NewPostExtraRepository(db *database.DB) PostExtraRepository {
	return &postExtraRepository{db: db}
}

func (r *postExtraRepository) UpsertByPostID(ctx context.Context, extras *model.PostExtra) error {
	var existed model.PostExtra
	err := database.Conn(ctx, r.db).Where("post_id = ?", extras.PostID).First(&existed).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return database.Conn(ctx, r.db).Create(extras).Error
	}
	extras.ID = existed.ID
	extras.CreatedAt = existed.CreatedAt
	return database.Conn(ctx, r.db).Save(extras).Error
}

func (r *postExtraRepository) DeleteByPostID(ctx context.Context, postID uint) error {
	return database.Conn(ctx, r.db).Where("post_id = ?", postID).Delete(&model.PostExtra{}).Error
}

func (r *postExtraRepository) GetByPostID(ctx context.Context, postID uint) (*model.PostExtra, error) {
	var extras model.PostExtra
	if err := database.Conn(ctx, r.db).Where("post_id = ?", postID).First(&extras).Error; err != nil {
		return nil, err
	}
	return &extras, nil
}

func (r *postExtraRepository) ListByPostIDs(ctx context.Context, postIDs []uint) ([]*model.PostExtra, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}
	var list []*model.PostExtra
	if err := database.Conn(ctx, r.db).Where("post_id IN ?", postIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
