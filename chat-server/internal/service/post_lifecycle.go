package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *PostService) IncrDiscussMessageCount(ctx context.Context, postID uint) error {
	if postID == 0 {
		return nil
	}
	if err := s.postRepo.IncrDiscussMessageCount(ctx, postID); err != nil {
		s.log.Warn(ctx, "增加讨论消息数失败", logger.Uint("post_id", postID), logger.Error(err))
		return err
	}
	return nil
}

func (s *PostService) DecrDiscussMessageCount(ctx context.Context, postID uint) error {
	if postID == 0 {
		return nil
	}
	if err := s.postRepo.DecrDiscussMessageCount(ctx, postID); err != nil {
		s.log.Warn(ctx, "减少讨论消息数失败", logger.Uint("post_id", postID), logger.Error(err))
		return err
	}
	return nil
}

func (s *PostService) Create(ctx context.Context, userID uint, req *types.CreatePostRequest) (uint, error) {
	if !validPostCategory(req.Category) {
		return 0, errs.Invalid("无效的动态分类")
	}
	placeName, lat, lng, err := s.resolvePostLocation(ctx, req.RoomID, req.PlaceName, req.Latitude, req.Longitude)
	if err != nil {
		return 0, err
	}
	fields, err := normalizePostWrite(
		req.Category, req.RoomID, req.Pinned,
		req.Title, req.Content, req.StartAt, req.EndAt, req.Extras,
	)
	if err != nil {
		return 0, err
	}
	feeDesc := ""
	if fields.Extras != nil {
		feeDesc = fields.Extras.FeeDescription
	}
	if err := assertClean(s.sensitive, fields.Title, fields.Content, placeName, feeDesc); err != nil {
		return 0, err
	}
	post := &model.Post{
		UserID:    userID,
		RoomID:    req.RoomID,
		Category:  req.Category,
		Pinned:    fields.Pinned,
		Title:     fields.Title,
		Content:   fields.Content,
		PlaceName: placeName,
		Latitude:  lat,
		Longitude: lng,
		StartAt:   fields.StartAt,
		EndAt:     fields.EndAt,
		Status:    model.PostStatusPending,
	}
	if len(req.Images) > 0 {
		images, err := normalizeMediaIDs(req.Images, 9)
		if err != nil {
			return 0, err
		}
		encoded, err := encodeMediaIDs(images)
		if err != nil {
			return 0, errs.Invalid("图片参数错误")
		}
		post.Images = encoded
	}
	if err := s.postRepo.Create(ctx, post); err != nil {
		return 0, fmt.Errorf("发布失败: %w", err)
	}
	if len(req.Images) > 0 && s.mediaSvc != nil {
		ids := parseMediaIDs(post.Images)
		if _, err := s.mediaSvc.BindMany(ctx, userID, ids, model.MediaKindImage, model.MediaBizPost, post.ID); err != nil {
			return 0, err
		}
	}
	if err := s.persistExtras(ctx, post.ID, fields.Extras); err != nil {
		return 0, fmt.Errorf("发布失败: %w", err)
	}
	return post.ID, nil
}

func (s *PostService) Update(ctx context.Context, userID uint, req *types.UpdatePostRequest) error {
	p, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if p.UserID != userID {
		return errs.Forbidden("无权编辑该动态")
	}
	if p.Status == model.PostStatusDisabled {
		return errs.Forbidden("动态已下架，无法编辑")
	}
	if p.Status == model.PostStatusPublished {
		return errs.Forbidden("动态已发布，无法编辑")
	}
	if !validPostCategory(req.Category) {
		return errs.Invalid("无效的动态分类")
	}
	placeName, lat, lng, err := s.resolvePostLocation(ctx, p.RoomID, req.PlaceName, req.Latitude, req.Longitude)
	if err != nil {
		return err
	}
	fields, err := normalizePostWrite(
		req.Category, p.RoomID, req.Pinned,
		req.Title, req.Content, req.StartAt, req.EndAt, req.Extras,
	)
	if err != nil {
		return err
	}
	feeDesc := ""
	if fields.Extras != nil {
		feeDesc = fields.Extras.FeeDescription
	}
	if err := assertClean(s.sensitive, fields.Title, fields.Content, placeName, feeDesc); err != nil {
		return err
	}

	p.Category = req.Category
	p.Pinned = fields.Pinned
	p.Title = fields.Title
	p.Content = fields.Content
	p.PlaceName = placeName
	p.Latitude = lat
	p.Longitude = lng
	p.StartAt = fields.StartAt
	p.EndAt = fields.EndAt
	var imageIDs []uint
	if s.mediaSvc != nil {
		ids, err := normalizeMediaIDs(req.Images, 9)
		if err != nil {
			return err
		}
		if _, err := s.mediaSvc.ReplaceBizMedia(ctx, userID, ids, model.MediaKindImage, model.MediaBizPost, p.ID); err != nil {
			return err
		}
		imageIDs = ids
	}
	encoded, err := encodeMediaIDs(imageIDs)
	if err != nil {
		return errs.Invalid("图片参数错误")
	}
	p.Images = encoded
	if p.Status == model.PostStatusRejected {
		p.Status = model.PostStatusPending
	}
	if err := s.postRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}

	if err := s.persistExtras(ctx, p.ID, fields.Extras); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	return nil
}

func (s *PostService) persistExtras(ctx context.Context, postID uint, extras *model.PostExtra) error {
	if extras == nil || !postExtrasHasContent(extras) {
		return s.extrasRepo.DeleteByPostID(ctx, postID)
	}
	extras.PostID = postID
	return s.extrasRepo.UpsertByPostID(ctx, extras)
}

func (s *PostService) resolvePostLocation(
	ctx context.Context,
	roomID uint,
	placeName string,
	lat, lng float64,
) (string, float64, float64, error) {
	hasClientLoc := !(lat == 0 && lng == 0) && placeName != ""

	if roomID > 0 {
		room, err := s.roomSvc.Get(ctx, roomID)
		if err != nil {
			return "", 0, 0, err
		}
		if hasClientLoc {
			return placeName, lat, lng, nil
		}
		if room.CenterLat == 0 && room.CenterLng == 0 {
			return "", 0, 0, errs.Invalid("聊天区位置无效")
		}
		name := room.Name
		if name == "" {
			name = "聊天区"
		}
		return name, room.CenterLat, room.CenterLng, nil
	}

	if lat == 0 && lng == 0 {
		return "", 0, 0, errs.Invalid("请提供有效地点坐标")
	}
	if placeName == "" {
		return "", 0, 0, errs.Invalid("请提供地点名称")
	}
	return placeName, lat, lng, nil
}

func (s *PostService) Withdraw(ctx context.Context, userID uint, postID uint) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if p.UserID != userID {
		return errs.Forbidden("无权操作该动态")
	}
	if p.Status == model.PostStatusDisabled {
		return errs.Forbidden("动态已下架")
	}
	p.Status = model.PostStatusDisabled
	p.AuditRemark = ""
	p.AuditedBy = 0
	p.AuditedAt = 0
	if err := s.postRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("下架失败: %w", err)
	}
	s.closePostChannel(ctx, postID, "动态已下架")
	return nil
}

func (s *PostService) DeleteMine(ctx context.Context, userID uint, postID uint) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if p.UserID != userID {
		return errs.Forbidden("无权操作该动态")
	}
	switch p.Status {
	case model.PostStatusPending, model.PostStatusRejected, model.PostStatusDisabled:

	default:
		return errs.Invalid("请先下架后再删除")
	}
	if err := s.postRepo.SoftDelete(ctx, postID); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	s.closePostChannel(ctx, postID, "动态已删除")
	return nil
}

func (s *PostService) HideJoined(ctx context.Context, userID uint, postID uint) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	item, err := s.signupRepo.GetByPostAndUser(ctx, postID, userID)
	if err != nil {
		return errs.NotFound("未找到报名记录")
	}
	if item.Status != model.SignupStatusConfirmed && item.Status != model.SignupStatusWaitlisted {
		return errs.Invalid("当前状态无法删除")
	}
	if item.HiddenAt > 0 {
		return nil
	}
	if !canHideJoined(p, item) {
		return errs.Invalid("活动进行中，请先取消报名")
	}
	item.HiddenAt = time.Now().Unix()
	if err := s.signupRepo.Update(ctx, item); err != nil {
		return fmt.Errorf("删除失败: %w", err)
	}
	return nil
}

func canHideJoined(p *model.Post, item *model.PostSignup) bool {
	if p.Status == model.PostStatusDisabled {
		return true
	}
	if item.CheckedInAt > 0 {
		return true
	}
	now := time.Now().Unix()
	if p.EndAt > 0 && now > p.EndAt {
		return true
	}
	return false
}

func (s *PostService) MessageCardSnapshot(ctx context.Context, postID uint) (uint, string, string, error) {
	if postID == 0 {
		return 0, "", "", errs.Invalid("动态无效")
	}
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", "", errs.NotFound("动态不存在或已下架")
		}
		return 0, "", "", fmt.Errorf("查询失败: %w", err)
	}
	if p.Status != model.PostStatusPublished {
		return 0, "", "", errs.NotFound("动态不存在或已下架")
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		title = "动态"
	}
	cover := ""
	ids := parseMediaIDs(p.Images)
	if len(ids) > 0 && s.mediaSvc != nil {
		urls := s.mediaSvc.ResolveURLList(ctx, ids[:1])
		if len(urls) > 0 {
			cover = urls[0]
		}
	}
	return p.ID, title, cover, nil
}
