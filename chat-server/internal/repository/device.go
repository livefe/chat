package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type DeviceRepository interface {
	Upsert(ctx context.Context, d *model.Device) error
	DeleteByUserAndDevice(ctx context.Context, userID uint, deviceID string) error
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteByToken(ctx context.Context, vendor, token string) error
	ListEnabledByUserID(ctx context.Context, userID uint) ([]*model.Device, error)
}

type deviceRepository struct {
	db *database.DB
}

func NewDeviceRepository(db *database.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Upsert(ctx context.Context, d *model.Device) error {
	now := time.Now().Unix()
	if d.LastActiveAt == 0 {
		d.LastActiveAt = now
	}
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"platform", "vendor", "push_token", "app_version", "os_version",
			"push_enabled", "last_active_at", "updated_at",
		}),
	}).Create(d).Error
}

func (r *deviceRepository) DeleteByUserAndDevice(ctx context.Context, userID uint, deviceID string) error {
	return database.Conn(ctx, r.db).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Delete(&model.Device{}).Error
}

func (r *deviceRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return database.Conn(ctx, r.db).Where("user_id = ?", userID).Delete(&model.Device{}).Error
}

func (r *deviceRepository) DeleteByToken(ctx context.Context, vendor, token string) error {
	return database.Conn(ctx, r.db).
		Where("vendor = ? AND push_token = ?", vendor, token).
		Delete(&model.Device{}).Error
}

func (r *deviceRepository) ListEnabledByUserID(ctx context.Context, userID uint) ([]*model.Device, error) {
	var list []*model.Device
	err := database.Conn(ctx, r.db).
		Where("user_id = ? AND push_enabled = ?", userID, true).
		Order("last_active_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}
