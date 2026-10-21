package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type SensitiveWordListFilter struct {
	Word     string
	Category string
	Status   *uint8
	Page     int
	PageSize int
}

type SensitiveWordRepository interface {
	Create(ctx context.Context, row *model.SensitiveWord) error
	Update(ctx context.Context, row *model.SensitiveWord) error
	GetByID(ctx context.Context, id uint) (*model.SensitiveWord, error)
	GetByWord(ctx context.Context, word string) (*model.SensitiveWord, error)
	List(ctx context.Context, f SensitiveWordListFilter) ([]*model.SensitiveWord, int64, error)
	ListEnabled(ctx context.Context) ([]*model.SensitiveWord, error)
	Delete(ctx context.Context, id uint) error
}

type sensitiveWordRepository struct {
	db *database.DB
}

func NewSensitiveWordRepository(db *database.DB) SensitiveWordRepository {
	return &sensitiveWordRepository{db: db}
}

func (r *sensitiveWordRepository) Create(ctx context.Context, row *model.SensitiveWord) error {
	return database.Conn(ctx, r.db).Create(row).Error
}

func (r *sensitiveWordRepository) Update(ctx context.Context, row *model.SensitiveWord) error {
	return database.Conn(ctx, r.db).Save(row).Error
}

func (r *sensitiveWordRepository) GetByID(ctx context.Context, id uint) (*model.SensitiveWord, error) {
	var row model.SensitiveWord
	if err := database.Conn(ctx, r.db).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *sensitiveWordRepository) GetByWord(ctx context.Context, word string) (*model.SensitiveWord, error) {
	var row model.SensitiveWord
	if err := database.Conn(ctx, r.db).Where("word = ?", word).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *sensitiveWordRepository) List(ctx context.Context, f SensitiveWordListFilter) ([]*model.SensitiveWord, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.SensitiveWord{})
	if f.Word != "" {
		db = db.Where("word LIKE ?", "%"+f.Word+"%")
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.SensitiveWord
	err := db.Order("id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error
	return list, total, err
}

func (r *sensitiveWordRepository) ListEnabled(ctx context.Context) ([]*model.SensitiveWord, error) {
	var list []*model.SensitiveWord
	err := database.Conn(ctx, r.db).
		Where("status = ?", 1).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *sensitiveWordRepository) Delete(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Unscoped().Delete(&model.SensitiveWord{}, id).Error
}
