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
	CreateBatch(ctx context.Context, rows []*model.SensitiveWord) error
	Update(ctx context.Context, row *model.SensitiveWord) error
	GetByID(ctx context.Context, id uint) (*model.SensitiveWord, error)
	List(ctx context.Context, f SensitiveWordListFilter) ([]*model.SensitiveWord, int64, error)
	ListAll(ctx context.Context) ([]*model.SensitiveWord, error)
	ListEnabled(ctx context.Context) ([]*model.SensitiveWord, error)
	Delete(ctx context.Context, id uint) error
	DeleteByIDs(ctx context.Context, ids []uint) error
	UpdateStatusByIDs(ctx context.Context, ids []uint, status uint8) error
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

func (r *sensitiveWordRepository) CreateBatch(ctx context.Context, rows []*model.SensitiveWord) error {
	if len(rows) == 0 {
		return nil
	}
	return database.Conn(ctx, r.db).CreateInBatches(rows, 200).Error
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

func (r *sensitiveWordRepository) ListAll(ctx context.Context) ([]*model.SensitiveWord, error) {
	var list []*model.SensitiveWord
	err := database.Conn(ctx, r.db).Select("id", "word").Order("id ASC").Find(&list).Error
	return list, err
}

func (r *sensitiveWordRepository) ListEnabled(ctx context.Context) ([]*model.SensitiveWord, error) {
	var list []*model.SensitiveWord
	err := database.Conn(ctx, r.db).
		Where("status = ?", model.SensitiveWordStatusEnabled).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *sensitiveWordRepository) Delete(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Unscoped().Delete(&model.SensitiveWord{}, id).Error
}

func (r *sensitiveWordRepository) DeleteByIDs(ctx context.Context, ids []uint) error {
	return database.Conn(ctx, r.db).Unscoped().Where("id IN ?", ids).Delete(&model.SensitiveWord{}).Error
}

func (r *sensitiveWordRepository) UpdateStatusByIDs(ctx context.Context, ids []uint, status uint8) error {
	return database.Conn(ctx, r.db).Model(&model.SensitiveWord{}).
		Where("id IN ?", ids).
		Update("status", status).Error
}
