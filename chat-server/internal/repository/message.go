package repository

import (
	"context"
	"errors"
	"time"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm"
)

type MessageListFilter struct {
	ID         uint
	UserID     uint
	PeerUserID uint
	RoomID     uint
	PostID     uint
	Scope      *uint8
	Keyword    string
	StartTime  string
	EndTime    string
	Page       int
	PageSize   int
}

type GeoBox struct {
	MinLat, MaxLat, MinLng, MaxLng float64
}

type UserRoomActivity struct {
	RoomID        uint
	MessageCount  int64
	LastMessageAt time.Time
}

type MessageRepository interface {
	Create(ctx context.Context, msg *model.Message) error
	GetByID(ctx context.Context, id uint) (*model.Message, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.Message, error)
	GetByIDsUnscoped(ctx context.Context, ids []uint) ([]*model.Message, error)
	GetByIDUnscoped(ctx context.Context, id uint) (*model.Message, error)
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, f MessageListFilter) ([]*model.Message, int64, error)
	ListContext(ctx context.Context, msg *model.Message, nearby *GeoBox, beforeID, afterID uint, limit int) ([]*model.Message, error)
	ListUserRoomActivity(ctx context.Context, userID uint, limit int) ([]UserRoomActivity, error)
	ListRoom(ctx context.Context, roomID uint, beforeID uint, limit int) ([]*model.Message, error)
	ListPost(ctx context.Context, postID uint, beforeID uint, limit int) ([]*model.Message, error)
	LatestVisiblePost(ctx context.Context, postID uint) (*model.Message, error)
	ListNearbyByBox(ctx context.Context, minLat, maxLat, minLng, maxLng float64, beforeID uint, limit int) ([]*model.Message, error)
	ListPrivate(ctx context.Context, userA, userB uint, clearedBefore uint, beforeID uint, limit int) ([]*model.Message, error)
	MaxPrivateIDBetween(ctx context.Context, userA, userB uint) (uint, error)
	MarkRecalled(ctx context.Context, id uint) error
	LastPrivateVisible(ctx context.Context, userA, userB uint) (*model.Message, error)
}

type messageRepository struct {
	db *database.DB
}

func NewMessageRepository(db *database.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *model.Message) error {
	return database.Conn(ctx, r.db).Create(msg).Error
}

func (r *messageRepository) GetByID(ctx context.Context, id uint) (*model.Message, error) {
	var msg model.Message
	if err := database.Conn(ctx, r.db).First(&msg, id).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) GetByIDs(ctx context.Context, ids []uint) ([]*model.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*model.Message
	if err := database.Conn(ctx, r.db).Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *messageRepository) GetByIDsUnscoped(ctx context.Context, ids []uint) ([]*model.Message, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*model.Message
	if err := database.Conn(ctx, r.db).Unscoped().Where("id IN ?", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *messageRepository) GetByIDUnscoped(ctx context.Context, id uint) (*model.Message, error) {
	var msg model.Message
	if err := database.Conn(ctx, r.db).Unscoped().First(&msg, id).Error; err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) Delete(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Delete(&model.Message{}, id).Error
}

func (r *messageRepository) applyAdminFilters(db *gorm.DB, f MessageListFilter) *gorm.DB {
	if f.ID > 0 {
		db = db.Where("id = ?", f.ID)
	}
	switch {
	case f.UserID > 0 && f.PeerUserID > 0:
		db = db.Where("scope = ?", model.MessageScopePrivate).
			Where(
				"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
				f.UserID, f.PeerUserID, f.PeerUserID, f.UserID,
			)
	case f.UserID > 0:
		db = db.Where("from_user_id = ? OR to_user_id = ?", f.UserID, f.UserID)
	case f.PeerUserID > 0:
		db = db.Where("scope = ?", model.MessageScopePrivate).
			Where("from_user_id = ? OR to_user_id = ?", f.PeerUserID, f.PeerUserID)
	}
	if f.RoomID > 0 {
		db = db.Where("room_id = ?", f.RoomID)
	}
	if f.PostID > 0 {
		db = db.Where("post_id = ?", f.PostID)
	}
	if f.Scope != nil {
		db = db.Where("scope = ?", *f.Scope)
	}
	if f.Keyword != "" {
		db = db.Where("content LIKE ?", "%"+f.Keyword+"%")
	}
	if f.StartTime != "" {
		db = db.Where("created_at >= ?", f.StartTime)
	}
	if f.EndTime != "" {
		db = db.Where("created_at <= ?", f.EndTime)
	}
	return db
}

func (r *messageRepository) List(ctx context.Context, f MessageListFilter) ([]*model.Message, int64, error) {
	db := r.applyAdminFilters(database.Conn(ctx, r.db).Unscoped().Model(&model.Message{}), f)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*model.Message
	if err := db.Order("id DESC").
		Offset((f.Page - 1) * f.PageSize).
		Limit(f.PageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// contextQuery 与 msg 处于同一会话的消息；附近消息按 nearby 区域圈定，缺少区域时只返回自身。
func (r *messageRepository) contextQuery(ctx context.Context, msg *model.Message, nearby *GeoBox) *gorm.DB {
	db := database.Conn(ctx, r.db).Unscoped().Model(&model.Message{})
	switch msg.Scope {
	case model.MessageScopeRoom:
		return db.Where("scope = ? AND room_id = ?", model.MessageScopeRoom, msg.RoomID)
	case model.MessageScopePost:
		return db.Where("scope = ? AND post_id = ?", model.MessageScopePost, msg.PostID)
	case model.MessageScopePrivate:
		return db.Where("scope = ?", model.MessageScopePrivate).
			Where(
				"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
				msg.FromUserID, msg.ToUserID, msg.ToUserID, msg.FromUserID,
			)
	case model.MessageScopeNearby:
		if nearby == nil {
			return db.Where("id = ?", msg.ID)
		}
		return db.Where("scope = ?", model.MessageScopeNearby).
			Where("latitude BETWEEN ? AND ?", nearby.MinLat, nearby.MaxLat).
			Where("longitude BETWEEN ? AND ?", nearby.MinLng, nearby.MaxLng)
	default:
		return db.Where("id = ?", msg.ID)
	}
}

func (r *messageRepository) ListContext(ctx context.Context, msg *model.Message, nearby *GeoBox, beforeID, afterID uint, limit int) ([]*model.Message, error) {
	if afterID > 0 {
		var after []*model.Message
		err := r.contextQuery(ctx, msg, nearby).Where("id > ?", afterID).Order("id ASC").Limit(limit).Find(&after).Error
		return after, err
	}

	pivot, beforeLimit := msg.ID+1, limit+1
	if beforeID > 0 {
		pivot, beforeLimit = beforeID, limit
	}
	var before []*model.Message
	if err := r.contextQuery(ctx, msg, nearby).Where("id < ?", pivot).Order("id DESC").Limit(beforeLimit).Find(&before).Error; err != nil {
		return nil, err
	}
	out := make([]*model.Message, 0, len(before)+limit)
	for i := len(before) - 1; i >= 0; i-- {
		out = append(out, before[i])
	}
	if beforeID > 0 {
		return out, nil
	}

	var after []*model.Message
	if err := r.contextQuery(ctx, msg, nearby).Where("id > ?", msg.ID).Order("id ASC").Limit(limit).Find(&after).Error; err != nil {
		return nil, err
	}
	return append(out, after...), nil
}

func (r *messageRepository) ListUserRoomActivity(ctx context.Context, userID uint, limit int) ([]UserRoomActivity, error) {
	var rows []UserRoomActivity
	err := database.Conn(ctx, r.db).Model(&model.Message{}).
		Select("room_id, COUNT(*) AS message_count, MAX(created_at) AS last_message_at").
		Where("scope = ? AND from_user_id = ? AND room_id > 0", model.MessageScopeRoom, userID).
		Group("room_id").
		Order("last_message_at DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *messageRepository) listByScopeFK(ctx context.Context, scope uint8, fkCol string, fkID, beforeID uint, limit int) ([]*model.Message, error) {
	db := database.Conn(ctx, r.db).Model(&model.Message{}).
		Where("scope = ? AND "+fkCol+" = ?", scope, fkID)
	if beforeID > 0 {
		db = db.Where("id < ?", beforeID)
	}
	var list []*model.Message
	if err := db.Order("id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *messageRepository) ListRoom(ctx context.Context, roomID uint, beforeID uint, limit int) ([]*model.Message, error) {
	return r.listByScopeFK(ctx, model.MessageScopeRoom, "room_id", roomID, beforeID, limit)
}

func (r *messageRepository) ListPost(ctx context.Context, postID uint, beforeID uint, limit int) ([]*model.Message, error) {
	return r.listByScopeFK(ctx, model.MessageScopePost, "post_id", postID, beforeID, limit)
}

func (r *messageRepository) LatestVisiblePost(ctx context.Context, postID uint) (*model.Message, error) {
	var msg model.Message
	err := database.Conn(ctx, r.db).
		Where("scope = ? AND post_id = ? AND recalled = 0", model.MessageScopePost, postID).
		Order("id DESC").
		First(&msg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *messageRepository) ListNearbyByBox(ctx context.Context, minLat, maxLat, minLng, maxLng float64, beforeID uint, limit int) ([]*model.Message, error) {
	db := database.Conn(ctx, r.db).Model(&model.Message{}).
		Where("scope = ?", model.MessageScopeNearby).
		Where("latitude BETWEEN ? AND ?", minLat, maxLat).
		Where("longitude BETWEEN ? AND ?", minLng, maxLng)
	if beforeID > 0 {
		db = db.Where("id < ?", beforeID)
	}
	var list []*model.Message
	if err := db.Order("id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *messageRepository) listPrivateDirection(
	ctx context.Context,
	fromID, toID uint,
	clearedBefore, beforeID uint,
	limit int,
	onlyVisible bool,
) ([]*model.Message, error) {
	db := database.Conn(ctx, r.db).Model(&model.Message{}).
		Where("scope = ? AND from_user_id = ? AND to_user_id = ?", model.MessageScopePrivate, fromID, toID)
	if clearedBefore > 0 {
		db = db.Where("id > ?", clearedBefore)
	}
	if beforeID > 0 {
		db = db.Where("id < ?", beforeID)
	}
	if onlyVisible {
		db = db.Where("recalled = 0")
	}
	var list []*model.Message
	if err := db.Order("id DESC").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func mergeMessagesByIDDesc(a, b []*model.Message, limit int) []*model.Message {
	out := make([]*model.Message, 0, min(limit, len(a)+len(b)))
	i, j := 0, 0
	for len(out) < limit && (i < len(a) || j < len(b)) {
		if j >= len(b) || (i < len(a) && a[i].ID >= b[j].ID) {
			out = append(out, a[i])
			i++
			continue
		}
		out = append(out, b[j])
		j++
	}
	return out
}

func (r *messageRepository) ListPrivate(ctx context.Context, userA, userB uint, clearedBefore uint, beforeID uint, limit int) ([]*model.Message, error) {
	ab, err := r.listPrivateDirection(ctx, userA, userB, clearedBefore, beforeID, limit, false)
	if err != nil {
		return nil, err
	}
	if userA == userB {
		return ab, nil
	}
	ba, err := r.listPrivateDirection(ctx, userB, userA, clearedBefore, beforeID, limit, false)
	if err != nil {
		return nil, err
	}
	return mergeMessagesByIDDesc(ab, ba, limit), nil
}

func (r *messageRepository) maxPrivateIDDirection(ctx context.Context, fromID, toID uint) (uint, error) {
	var maxID *uint
	err := database.Conn(ctx, r.db).Model(&model.Message{}).
		Where("scope = ? AND from_user_id = ? AND to_user_id = ?", model.MessageScopePrivate, fromID, toID).
		Select("MAX(id)").Scan(&maxID).Error
	if err != nil {
		return 0, err
	}
	if maxID == nil {
		return 0, nil
	}
	return *maxID, nil
}

func (r *messageRepository) MaxPrivateIDBetween(ctx context.Context, userA, userB uint) (uint, error) {
	ab, err := r.maxPrivateIDDirection(ctx, userA, userB)
	if err != nil {
		return 0, err
	}
	ba, err := r.maxPrivateIDDirection(ctx, userB, userA)
	if err != nil {
		return 0, err
	}
	if ab >= ba {
		return ab, nil
	}
	return ba, nil
}

func (r *messageRepository) MarkRecalled(ctx context.Context, id uint) error {
	return database.Conn(ctx, r.db).Model(&model.Message{}).
		Where("id = ?", id).
		Update("recalled", 1).Error
}

func (r *messageRepository) LastPrivateVisible(ctx context.Context, userA, userB uint) (*model.Message, error) {
	ab, err := r.listPrivateDirection(ctx, userA, userB, 0, 0, 1, true)
	if err != nil {
		return nil, err
	}
	ba, err := r.listPrivateDirection(ctx, userB, userA, 0, 0, 1, true)
	if err != nil {
		return nil, err
	}
	merged := mergeMessagesByIDDesc(ab, ba, 1)
	if len(merged) == 0 {
		return nil, nil
	}
	return merged[0], nil
}
