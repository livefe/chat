package repository

import (
	"context"
	"math"

	"app/internal/model"
	"app/pkg/database"
)

type RoomListFilter struct {
	ID        *uint
	UserID    *uint
	Name      string
	Status    *uint8
	Source    *uint8
	PlaceType *uint8
	Page      int
	PageSize  int
}

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	Update(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id uint) (*model.Room, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.Room, error)
	List(ctx context.Context, f RoomListFilter) ([]*model.Room, int64, error)
	ListNearCenter(ctx context.Context, lat, lng, radiusM float64, excludeID uint, statuses []uint8, limit int) ([]*model.Room, error)
	CountByStatus(ctx context.Context, status uint8) (int64, error)
}

type roomRepository struct {
	db *database.DB
}

func NewRoomRepository(db *database.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(ctx context.Context, room *model.Room) error {
	return database.Conn(ctx, r.db).Create(room).Error
}

func (r *roomRepository) Update(ctx context.Context, room *model.Room) error {
	return database.Conn(ctx, r.db).Save(room).Error
}

func (r *roomRepository) GetByID(ctx context.Context, id uint) (*model.Room, error) {
	var room model.Room
	if err := database.Conn(ctx, r.db).First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.Room, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rooms []*model.Room
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) List(ctx context.Context, f RoomListFilter) ([]*model.Room, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.Room{})

	if f.ID != nil {
		db = db.Where("id = ?", *f.ID)
	}
	if f.UserID != nil {
		db = db.Where("user_id = ?", *f.UserID)
	}
	if f.Name != "" {
		db = db.Where("name LIKE ?", "%"+f.Name+"%")
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.Source != nil {
		db = db.Where("source = ?", *f.Source)
	}
	if f.PlaceType != nil {
		db = db.Where("place_type = ?", *f.PlaceType)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rooms []*model.Room
	if err := db.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&rooms).Error; err != nil {
		return nil, 0, err
	}
	return rooms, total, nil
}

func (r *roomRepository) ListNearCenter(ctx context.Context, lat, lng, radiusM float64, excludeID uint, statuses []uint8, limit int) ([]*model.Room, error) {
	deltaLat := radiusM / 111000.0
	cosLat := math.Cos(lat * math.Pi / 180)
	if cosLat < 0.2 {
		cosLat = 0.2
	}
	deltaLng := radiusM / (111000.0 * cosLat)

	db := database.Conn(ctx, r.db).Model(&model.Room{}).
		Where("center_lat BETWEEN ? AND ?", lat-deltaLat, lat+deltaLat).
		Where("center_lng BETWEEN ? AND ?", lng-deltaLng, lng+deltaLng)
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	if len(statuses) > 0 {

		statusArgs := make([]int, len(statuses))
		for i, s := range statuses {
			statusArgs[i] = int(s)
		}
		db = db.Where("status IN ?", statusArgs)
	}

	var rooms []*model.Room
	if err := db.Limit(limit * 3).Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Room{}).Where("status = ?", status).Count(&n).Error
	return n, err
}
