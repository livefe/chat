package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

const (
	AppPlatformAndroid = "android"
	AppPlatformIOS     = "ios"
	AppPlatformH5      = "h5"
)

type AppReleaseRepository interface {
	Create(ctx context.Context, row *model.AppRelease) error
	GetByID(ctx context.Context, id uint) (*model.AppRelease, error)
	GetByPlatformVersion(ctx context.Context, platform, version string) (*model.AppRelease, error)
	GetCurrent(ctx context.Context, platform string) (*model.AppRelease, error)
	ListCurrent(ctx context.Context) ([]*model.AppRelease, error)
	List(ctx context.Context, platform string, page, pageSize int) ([]*model.AppRelease, int64, error)
	SetCurrent(ctx context.Context, platform string, id uint) error
	ClearCurrent(ctx context.Context, id uint) error
	DeleteNotCurrent(ctx context.Context, id uint) error
}

type appReleaseRepository struct {
	db *database.DB
}

func NewAppReleaseRepository(db *database.DB) AppReleaseRepository {
	return &appReleaseRepository{db: db}
}

func (r *appReleaseRepository) Create(ctx context.Context, row *model.AppRelease) error {
	return database.Conn(ctx, r.db).Create(row).Error
}

func (r *appReleaseRepository) GetByID(ctx context.Context, id uint) (*model.AppRelease, error) {
	var row model.AppRelease
	if err := database.Conn(ctx, r.db).First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *appReleaseRepository) GetByPlatformVersion(ctx context.Context, platform, version string) (*model.AppRelease, error) {
	var row model.AppRelease
	err := database.Conn(ctx, r.db).
		Where("platform = ? AND version = ?", platform, version).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *appReleaseRepository) GetCurrent(ctx context.Context, platform string) (*model.AppRelease, error) {
	var row model.AppRelease
	err := database.Conn(ctx, r.db).
		Where("platform = ? AND is_current = 1", platform).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *appReleaseRepository) ListCurrent(ctx context.Context) ([]*model.AppRelease, error) {
	var list []*model.AppRelease
	err := database.Conn(ctx, r.db).
		Where("is_current = 1").
		Order("platform ASC").
		Find(&list).Error
	return list, err
}

func (r *appReleaseRepository) List(ctx context.Context, platform string, page, pageSize int) ([]*model.AppRelease, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.AppRelease{})
	if platform != "" {
		db = db.Where("platform = ?", platform)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.AppRelease
	err := db.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

// DeleteNotCurrent 删除非当前对外的版本记录；记录不存在或为当前版本时返回 gorm.ErrRecordNotFound。
func (r *appReleaseRepository) DeleteNotCurrent(ctx context.Context, id uint) error {
	res := database.Conn(ctx, r.db).Where("id = ? AND is_current = 0", id).Delete(&model.AppRelease{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *appReleaseRepository) SetCurrent(ctx context.Context, platform string, id uint) error {
	return database.Conn(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AppRelease{}).
			Where("platform = ? AND is_current = 1", platform).
			Update("is_current", 0).Error; err != nil {
			return err
		}
		res := tx.Model(&model.AppRelease{}).
			Where("id = ? AND platform = ?", id, platform).
			Update("is_current", 1)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *appReleaseRepository) ClearCurrent(ctx context.Context, id uint) error {
	res := database.Conn(ctx, r.db).Model(&model.AppRelease{}).
		Where("id = ? AND is_current = 1", id).
		Update("is_current", 0)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
