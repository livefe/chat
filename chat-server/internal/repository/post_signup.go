package repository

import (
	"context"
	"strings"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type PostSignupRepository interface {
	Create(ctx context.Context, item *model.PostSignup) error
	Update(ctx context.Context, item *model.PostSignup) error
	GetByPostAndUser(ctx context.Context, postID uint, userID uint) (*model.PostSignup, error)
	GetByCheckInToken(ctx context.Context, token string) (*model.PostSignup, error)
	CountByPostAndStatus(ctx context.Context, postID uint, status uint8) (int64, error)
	CountListByPost(ctx context.Context, postID uint) (SignupListCounts, error)
	CountSummaryByPostIDs(ctx context.Context, postIDs []uint) (map[uint]SignupSummary, error)
	ListByUserAndPostIDs(ctx context.Context, userID uint, postIDs []uint) (map[uint]*model.PostSignup, error)
	GetFirstWaitlisted(ctx context.Context, postID uint) (*model.PostSignup, error)
	ListByPost(ctx context.Context, f SignupListFilter) ([]*model.PostSignup, int64, error)
	ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]*model.PostSignup, int64, error)
}

const (
	SignupSortCreatedAt   = "createdAt"
	SignupSortCheckedInAt = "checkedInAt"
)

type SignupListFilter struct {
	PostID             uint
	Status             *uint8
	CheckIn            *uint8
	UserID             *uint
	Keyword            string
	KeywordMatchMobile bool
	SortField          string
	SortDesc           bool
	Page               int
	PageSize           int
}

type SignupSummary struct {
	Confirmed  int64
	Waitlisted int64
}

type SignupListCounts struct {
	Total        int64 `gorm:"column:total"`
	Confirmed    int64 `gorm:"column:confirmed"`
	Waitlisted   int64 `gorm:"column:waitlisted"`
	Cancelled    int64 `gorm:"column:cancelled"`
	NotCheckedIn int64 `gorm:"column:not_checked_in"`
	CheckedIn    int64 `gorm:"column:checked_in"`
}

type postSignupRepository struct {
	db *database.DB
}

func NewPostSignupRepository(db *database.DB) PostSignupRepository {
	return &postSignupRepository{db: db}
}

func (r *postSignupRepository) Create(ctx context.Context, item *model.PostSignup) error {
	return database.Conn(ctx, r.db).Create(item).Error
}

func (r *postSignupRepository) Update(ctx context.Context, item *model.PostSignup) error {
	return database.Conn(ctx, r.db).Save(item).Error
}

func (r *postSignupRepository) GetByPostAndUser(ctx context.Context, postID uint, userID uint) (*model.PostSignup, error) {
	var item model.PostSignup
	if err := database.Conn(ctx, r.db).Where("post_id = ? AND user_id = ?", postID, userID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postSignupRepository) GetByCheckInToken(ctx context.Context, token string) (*model.PostSignup, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var item model.PostSignup
	if err := database.Conn(ctx, r.db).
		Where("check_in_token = ? AND check_in_token <> '' AND status = ?", token, model.SignupStatusConfirmed).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *postSignupRepository) CountByPostAndStatus(ctx context.Context, postID uint, status uint8) (int64, error) {
	var count int64
	if err := database.Conn(ctx, r.db).Model(&model.PostSignup{}).
		Where("post_id = ? AND status = ?", postID, status).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *postSignupRepository) CountListByPost(ctx context.Context, postID uint) (SignupListCounts, error) {
	var counts SignupListCounts
	err := database.Conn(ctx, r.db).Model(&model.PostSignup{}).
		Select(`COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS confirmed,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS waitlisted,
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0) AS cancelled,
			COALESCE(SUM(CASE WHEN status = ? AND checked_in_at = 0 THEN 1 ELSE 0 END), 0) AS not_checked_in,
			COALESCE(SUM(CASE WHEN status = ? AND checked_in_at > 0 THEN 1 ELSE 0 END), 0) AS checked_in`,
			model.SignupStatusConfirmed,
			model.SignupStatusWaitlisted,
			model.SignupStatusCancelled,
			model.SignupStatusConfirmed,
			model.SignupStatusConfirmed,
		).
		Where("post_id = ?", postID).
		Scan(&counts).Error
	if err != nil {
		return SignupListCounts{}, err
	}
	return counts, nil
}

func (r *postSignupRepository) CountSummaryByPostIDs(ctx context.Context, postIDs []uint) (map[uint]SignupSummary, error) {
	result := make(map[uint]SignupSummary)
	if len(postIDs) == 0 {
		return result, nil
	}
	type row struct {
		PostID uint
		Status uint8
		Cnt    int64
	}
	var rows []row
	err := database.Conn(ctx, r.db).Model(&model.PostSignup{}).
		Select("post_id, status, COUNT(*) as cnt").
		Where("post_id IN ? AND status IN ?", postIDs, []int{
			int(model.SignupStatusConfirmed),
			int(model.SignupStatusWaitlisted),
		}).
		Group("post_id, status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		summary := result[item.PostID]
		switch item.Status {
		case model.SignupStatusConfirmed:
			summary.Confirmed = item.Cnt
		case model.SignupStatusWaitlisted:
			summary.Waitlisted = item.Cnt
		}
		result[item.PostID] = summary
	}
	return result, nil
}

func (r *postSignupRepository) ListByUserAndPostIDs(ctx context.Context, userID uint, postIDs []uint) (map[uint]*model.PostSignup, error) {
	result := make(map[uint]*model.PostSignup)
	if userID == 0 || len(postIDs) == 0 {
		return result, nil
	}
	var list []*model.PostSignup
	if err := database.Conn(ctx, r.db).
		Where("user_id = ? AND post_id IN ?", userID, postIDs).
		Find(&list).Error; err != nil {
		return nil, err
	}
	for _, item := range list {
		result[item.PostID] = item
	}
	return result, nil
}

func (r *postSignupRepository) GetFirstWaitlisted(ctx context.Context, postID uint) (*model.PostSignup, error) {
	var item model.PostSignup
	if err := database.Conn(ctx, r.db).Where("post_id = ? AND status = ?", postID, model.SignupStatusWaitlisted).
		Order("created_at ASC").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func validSignupCheckIn(v uint8) bool {
	return v == model.SignupCheckInNotChecked || v == model.SignupCheckInChecked
}

func (r *postSignupRepository) ListByPost(ctx context.Context, f SignupListFilter) ([]*model.PostSignup, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.PostSignup{}).Where("post_id = ?", f.PostID)
	if f.CheckIn != nil && validSignupCheckIn(*f.CheckIn) {
		db = db.Where("status = ?", model.SignupStatusConfirmed)
		switch *f.CheckIn {
		case model.SignupCheckInNotChecked:
			db = db.Where("checked_in_at = 0")
		case model.SignupCheckInChecked:
			db = db.Where("checked_in_at > 0")
		}
	} else if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.UserID != nil {
		db = db.Where("user_id = ?", *f.UserID)
	}
	if f.Keyword != "" {
		like := contentLikePattern(f.Keyword)
		if f.KeywordMatchMobile {
			db = db.Where("user_id IN (SELECT id FROM `user` WHERE nickname LIKE ? OR mobile LIKE ?)", like, like)
		} else {
			db = db.Where("user_id IN (SELECT id FROM `user` WHERE nickname LIKE ?)", like)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.PostSignup
	if err := db.Order(signupListOrder(f.SortField, f.SortDesc)).
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func signupListOrder(field string, desc bool) string {
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	switch field {
	case SignupSortCreatedAt:
		return "created_at " + dir + ", id " + dir
	case SignupSortCheckedInAt:
		return "checked_in_at = 0 ASC, checked_in_at " + dir + ", id " + dir
	default:
		return "status ASC, created_at ASC, id ASC"
	}
}

func (r *postSignupRepository) ListByUser(ctx context.Context, userID uint, page, pageSize int) ([]*model.PostSignup, int64, error) {
	db := database.Conn(ctx, r.db).Model(&model.PostSignup{}).Where("user_id = ?", userID)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.PostSignup
	if err := db.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
