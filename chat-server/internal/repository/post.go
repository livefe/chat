package repository

import (
	"context"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type PostListFilter struct {
	Status         *uint8
	UserID         *uint
	RoomID         *uint
	Category       string
	Keyword        string
	OnlyPublicNow  bool
	ExcludeUserIDs []uint
	Page           int
	PageSize       int
}

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	Update(ctx context.Context, post *model.Post) error
	SoftDelete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.Post, error)
	List(ctx context.Context, f PostListFilter) ([]*model.Post, int64, error)
	ListJoinedByUser(ctx context.Context, userID uint, f PostListFilter) ([]*model.Post, int64, error)
	CountByStatus(ctx context.Context, status uint8) (int64, error)
	CountGroupedByRoom(ctx context.Context, status *uint8, category string) ([]PostRoomCount, error)
	ListIndependentForMap(ctx context.Context, status *uint8, category string, limit int) ([]*model.Post, error)
	IncrDiscussMessageCount(ctx context.Context, postID uint) error
	DecrDiscussMessageCount(ctx context.Context, postID uint) error
}

type PostRoomCount struct {
	RoomID uint  `gorm:"column:room_id"`
	Count  int64 `gorm:"column:cnt"`
}

type postRepository struct {
	db *database.DB
}

func NewPostRepository(db *database.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {

	return database.Conn(ctx, r.db).Select("*").Create(post).Error
}

func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	return database.Conn(ctx, r.db).Save(post).Error
}

func (r *postRepository) SoftDelete(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Delete(&model.Post{}, id).Error
}

func (r *postRepository) GetByID(ctx context.Context, id uint) (*model.Post, error) {
	var post model.Post
	if err := database.Conn(ctx, r.db).First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) List(ctx context.Context, f PostListFilter) ([]*model.Post, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.Post{})
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.UserID != nil {
		db = db.Where("user_id = ?", *f.UserID)
	}
	if len(f.ExcludeUserIDs) > 0 {
		db = db.Where("user_id NOT IN ?", f.ExcludeUserIDs)
	}
	if f.RoomID != nil {
		db = db.Where("room_id = ?", *f.RoomID)
	}
	if f.Category != "" {
		db = db.Where("category = ?", f.Category)
	}
	if f.Keyword != "" {
		kw := "%" + f.Keyword + "%"
		db = db.Where("title LIKE ? OR place_name LIKE ?", kw, kw)
	}
	if f.OnlyPublicNow {
		db = db.Where("status = ?", model.PostStatusPublished).
			Where("end_at = 0 OR end_at >= ?", time.Now().Unix())
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at DESC"
	if f.RoomID != nil && *f.RoomID > 0 {
		order = "pinned DESC, created_at DESC"
	}

	var list []*model.Post
	if err := db.Order(order).
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *postRepository) ListJoinedByUser(ctx context.Context, userID uint, f PostListFilter) ([]*model.Post, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.Post{}).
		Joins("JOIN post_signup s ON s.post_id = post.id").
		Where("s.user_id = ? AND s.status IN ? AND s.hidden_at = 0", userID, []int{
			int(model.SignupStatusConfirmed),
			int(model.SignupStatusWaitlisted),
		}).
		Group("post.id")

	if f.Status != nil {
		db = db.Where("post.status = ?", *f.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.Post
	if err := db.Order("post.created_at DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *postRepository) CountGroupedByRoom(ctx context.Context, status *uint8, category string) ([]PostRoomCount, error) {
	db := database.Conn(ctx, r.db).Model(&model.Post{}).
		Select("room_id, COUNT(*) AS cnt").
		Where("room_id > 0")
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var rows []PostRoomCount
	if err := db.Group("room_id").Order("cnt DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *postRepository) ListIndependentForMap(ctx context.Context, status *uint8, category string, limit int) ([]*model.Post, error) {
	db := database.Conn(ctx, r.db).Model(&model.Post{}).
		Where("room_id = 0").
		Where("(latitude != 0 OR longitude != 0)")
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	var list []*model.Post
	if err := db.Order("created_at DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *postRepository) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	var n int64
	err := database.Conn(ctx, r.db).Model(&model.Post{}).Where("status = ?", status).Count(&n).Error
	return n, err
}

func (r *postRepository) IncrDiscussMessageCount(ctx context.Context, postID uint) error {
	if postID == 0 {
		return nil
	}
	return database.Conn(ctx, r.db).Model(&model.Post{}).
		Where("id = ?", postID).
		UpdateColumn("discuss_message_count", gorm.Expr("discuss_message_count + 1")).Error
}

func (r *postRepository) DecrDiscussMessageCount(ctx context.Context, postID uint) error {
	if postID == 0 {
		return nil
	}
	return database.Conn(ctx, r.db).Model(&model.Post{}).
		Where("id = ? AND discuss_message_count > 0", postID).
		UpdateColumn("discuss_message_count", gorm.Expr("discuss_message_count - 1")).Error
}
