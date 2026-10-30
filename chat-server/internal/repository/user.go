package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/database"

	"gorm.io/gorm"
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

type UserUpdate struct {
	Mobile        *string
	Nickname      *string
	AvatarMediaID *uint
	Status        *uint8
	MutedUntil    *int64
}

func (u UserUpdate) columns() map[string]any {
	cols := make(map[string]any, 5)
	if u.Mobile != nil {
		cols["mobile"] = *u.Mobile
	}
	if u.Nickname != nil {
		cols["nickname"] = *u.Nickname
	}
	if u.AvatarMediaID != nil {
		cols["avatar_media_id"] = *u.AvatarMediaID
	}
	if u.Status != nil {
		cols["status"] = *u.Status
	}
	if u.MutedUntil != nil {
		cols["muted_until"] = *u.MutedUntil
	}
	return cols
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, id uint, u UserUpdate) error
	Deactivate(ctx context.Context, id uint, anonymizedMobile string) error
	Invalidate(ctx context.Context, id uint)

	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByIDUnscoped(ctx context.Context, id uint) (*model.User, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.User, error)
	GetByMobile(ctx context.Context, mobile string) (*model.User, error)
	List(ctx context.Context, f UserListFilter) ([]*model.User, int64, error)

	GetCached(ctx context.Context, id uint) (*model.User, error)
	GetManyCached(ctx context.Context, ids []uint) ([]*model.User, error)
}

type userRepository struct {
	db    *database.DB
	cache *cache.Cache
}

func NewUserRepository(db *database.DB, c *cache.Cache) UserRepository {
	return &userRepository{db: db, cache: c}
}

func userCacheKey(id uint) string {
	return fmt.Sprintf(types.CacheUserByIDKey, id)
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := database.Conn(ctx, r.db).Create(user).Error; err != nil {
		return err
	}
	r.Invalidate(ctx, user.ID)
	return nil
}

func (r *userRepository) Update(ctx context.Context, id uint, u UserUpdate) error {
	cols := u.columns()
	if len(cols) == 0 {
		return nil
	}
	if err := database.Conn(ctx, r.db).Model(&model.User{}).Where("id = ?", id).Updates(cols).Error; err != nil {
		return err
	}
	r.Invalidate(ctx, id)
	return nil
}

func (r *userRepository) Deactivate(ctx context.Context, id uint, anonymizedMobile string) error {
	res := database.Conn(ctx, r.db).Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{
		"mobile":          anonymizedMobile,
		"nickname":        model.DeactivatedNickname,
		"avatar_media_id": 0,
		"status":          model.UserStatusDisabled,
		"deleted_at":      time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	r.Invalidate(ctx, id)
	return nil
}

func (r *userRepository) Invalidate(ctx context.Context, id uint) {
	if id == 0 {
		return
	}
	_ = r.cache.Del(ctx, userCacheKey(id)).Err()
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
		like := "%" + f.Mobile + "%"
		deactivated := database.Conn(ctx, r.db).Model(&model.UserDeactivationLog{}).
			Select("user_id").Where("mobile LIKE ?", like)
		db = db.Unscoped().Where("((deleted_at IS NULL AND mobile LIKE ?) OR id IN (?))", like, deactivated)
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

func (r *userRepository) GetCached(ctx context.Context, id uint) (*model.User, error) {
	users, err := r.GetManyCached(ctx, []uint{id})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return users[0], nil
}

func (r *userRepository) GetManyCached(ctx context.Context, ids []uint) ([]*model.User, error) {
	uniq := uniqueNonZero(ids)
	if len(uniq) == 0 {
		return nil, nil
	}
	keys := make([]string, len(uniq))
	for i, id := range uniq {
		keys[i] = userCacheKey(id)
	}

	out := make([]*model.User, 0, len(uniq))
	miss := uniq
	if vals, err := r.cache.MGet(ctx, keys...).Result(); err == nil {
		miss = make([]uint, 0)
		for i, v := range vals {
			raw, ok := v.(string)
			if !ok {
				miss = append(miss, uniq[i])
				continue
			}
			if raw == cache.NullSentinel {
				continue
			}
			var user model.User
			if err := json.Unmarshal([]byte(raw), &user); err != nil {
				miss = append(miss, uniq[i])
				continue
			}
			out = append(out, &user)
		}
	}
	if len(miss) == 0 {
		return out, nil
	}

	var loaded []*model.User
	if err := r.db.WithContext(ctx).Where("id IN ?", miss).Find(&loaded).Error; err != nil {
		return nil, err
	}
	found := make(map[uint]struct{}, len(loaded))
	pipe := r.cache.Pipeline()
	for _, u := range loaded {
		found[u.ID] = struct{}{}
		out = append(out, u)
		if b, err := json.Marshal(u); err == nil {
			pipe.Set(ctx, userCacheKey(u.ID), b, types.Expire1h)
		}
	}
	for _, id := range miss {
		if _, ok := found[id]; !ok {
			pipe.Set(ctx, userCacheKey(id), cache.NullSentinel, types.Expire5m)
		}
	}
	_, _ = pipe.Exec(ctx)
	return out, nil
}

func uniqueNonZero(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
