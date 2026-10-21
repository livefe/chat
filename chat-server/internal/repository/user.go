package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"
)

type UserListFilter struct {
	ID        uint
	Mobile    string
	Nickname  string
	Status    *uint8
	Muted     *bool
	StartTime string
	EndTime   string
	Page      int
	PageSize  int
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Deactivate(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByIDUnscoped(ctx context.Context, id uint) (*model.User, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.User, error)
	GetByMobile(ctx context.Context, mobile string) (*model.User, error)
	List(ctx context.Context, f UserListFilter) ([]*model.User, int64, error)
}

type userRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return database.Conn(ctx, r.db).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return database.Conn(ctx, r.db).Save(user).Error
}

func (r *userRepository) Deactivate(ctx context.Context, user *model.User) error {
	if err := database.Conn(ctx, r.db).Save(user).Error; err != nil {
		return err
	}
	return database.Conn(ctx, r.db).Delete(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := database.Conn(ctx, r.db).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIDUnscoped(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := database.Conn(ctx, r.db).Unscoped().First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var users []*model.User
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetByMobile(ctx context.Context, mobile string) (*model.User, error) {
	var user model.User
	if err := database.Conn(ctx, r.db).Where("mobile = ?", mobile).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, f UserListFilter) ([]*model.User, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.User{})

	if f.ID > 0 {
		db = db.Unscoped().Where("id = ?", f.ID)
	}

	if f.Mobile != "" {
		db = db.Where("mobile LIKE ?", "%"+f.Mobile+"%")
	}
	if f.Nickname != "" {
		db = db.Where("nickname LIKE ?", "%"+f.Nickname+"%")
	}
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.Muted != nil && *f.Muted {
		nowSec := time.Now().Unix()
		db = db.Where("muted_until = ? OR muted_until > ?", model.PermanentMuteUntil, nowSec)
	}
	if f.StartTime != "" {
		db = db.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime != "" {
		db = db.Where("created_at <= ?", f.EndTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*model.User
	if err := db.Order("created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
