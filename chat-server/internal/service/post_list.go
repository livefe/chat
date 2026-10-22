package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/utils"

	"gorm.io/gorm"
)

func (s *PostService) ReportTarget(ctx context.Context, reporterID, targetID uint) (uint, types.ReportContextSnapshot, error) {
	post, err := s.postRepo.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, types.ReportContextSnapshot{}, errs.NotFound("动态不存在")
		}
		return 0, types.ReportContextSnapshot{}, fmt.Errorf("查询失败: %w", err)
	}
	if post.UserID == reporterID {
		return 0, types.ReportContextSnapshot{}, errs.Forbidden("不能举报自己的动态")
	}
	if post.Status != model.PostStatusPublished {
		return 0, types.ReportContextSnapshot{}, errs.Forbidden("动态不可举报")
	}
	var images []string
	if s.mediaSvc != nil {
		images = s.mediaSvc.ResolveURLList(ctx, parseMediaIDs(post.Images))
	}
	snapshot := types.ReportContextSnapshot{
		PostTitle:   post.Title,
		PostContent: post.Content,
		PostImages:  images,
	}
	profiles, err := s.userSvc.GetProfiles(ctx, []uint{post.UserID})
	if err != nil {
		return 0, types.ReportContextSnapshot{}, fmt.Errorf("查询失败: %w", err)
	}
	if profile, ok := profiles[post.UserID]; ok {
		snapshot.Nickname = profile.Nickname
		snapshot.Avatar = profile.AvatarURL
	}
	return post.UserID, snapshot, nil
}

func (s *PostService) ListPublic(ctx context.Context, viewerID uint, req *types.PostListRequest) (*types.DiscoverFeedResponse, error) {
	filter := repository.PostListFilter{
		Page:          req.Page,
		PageSize:      req.PageSize,
		OnlyPublicNow: true,
		Category:      req.Category,
	}

	if req.RoomID != nil && *req.RoomID > 0 {
		filter.RoomID = req.RoomID
		filter.PinnedFirst = true
	}
	if viewerID > 0 && s.userSvc != nil {
		if ids, err := s.userSvc.ListBlockedEitherPeerIDs(ctx, viewerID); err == nil {
			filter.ExcludeUserIDs = ids
		}
	}
	list, total, err := s.postRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.buildPostListItems(ctx, list, req.Latitude, req.Longitude)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	hasLocation := req.Latitude != 0 || req.Longitude != 0

	if s.adSvc != nil && (req.RoomID == nil || *req.RoomID == 0) {
		return s.adSvc.MergeDiscoverFeed(ctx, viewerID, items, total, req.Page, req.PageSize, req.Latitude, req.Longitude, hasLocation), nil
	}
	out := &types.DiscoverFeedResponse{Total: total, List: make([]types.FeedItem, 0, len(items))}
	for i := range items {
		p := items[i]
		out.List = append(out.List, types.FeedItem{Type: types.FeedItemTypePost, Post: &p})
	}
	return out, nil
}

func (s *PostService) ListMine(ctx context.Context, userID uint, req *types.MyPostListRequest) (*types.PostListResponse, error) {
	list, total, err := s.postRepo.List(ctx, repository.PostListFilter{
		UserID:   &userID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	resp, err := s.postListResponse(ctx, list, total, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}
	for i, p := range list {
		resp.List[i].AuditRemark = ownerAuditRemark(p)
	}
	return resp, nil
}

func (s *PostService) ListJoined(ctx context.Context, userID uint, req *types.MyPostListRequest) (*types.PostListResponse, error) {
	list, total, err := s.postRepo.ListJoinedByUser(ctx, userID, repository.PostListFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.buildPostListItems(ctx, list, req.Latitude, req.Longitude)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if err := s.attachMySignups(ctx, userID, items); err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.PostListResponse{Total: total, List: items}, nil
}

func (s *PostService) attachMySignups(ctx context.Context, userID uint, items []types.PostItem) error {
	if userID == 0 || len(items) == 0 {
		return nil
	}
	postIDs := make([]uint, len(items))
	for i := range items {
		postIDs[i] = items[i].ID
	}
	signupMap, err := s.signupRepo.ListByUserAndPostIDs(ctx, userID, postIDs)
	if err != nil {
		return fmt.Errorf("调用 ListByUserAndPostIDs 失败: %w", err)
	}
	for i := range items {
		signup, ok := signupMap[items[i].ID]
		if !ok {
			continue
		}
		items[i].MySignupStatus = signup.Status
		items[i].MyCheckedInAt = signup.CheckedInAt
	}
	return nil
}

func (s *PostService) ListPublicByUser(ctx context.Context, viewerID, userID uint, req *types.UserPostListRequest) (*types.PostListResponse, error) {
	if viewerID > 0 && s.userSvc != nil {
		blocked, err := s.userSvc.IsBlockedEither(ctx, viewerID, userID)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		if blocked {
			return &types.PostListResponse{Total: 0, List: []types.PostItem{}}, nil
		}
	}
	profiles, err := s.userSvc.GetProfiles(ctx, []uint{userID})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	publisher, ok := profiles[userID]
	if !ok {
		return nil, errs.NotFound("用户不存在")
	}
	if publisher.Nickname == model.DeactivatedNickname || publisher.Status != model.UserStatusEnabled {
		return &types.PostListResponse{Total: 0, List: []types.PostItem{}}, nil
	}
	published := model.PostStatusPublished
	return s.listPosts(ctx, repository.PostListFilter{
		UserID:        &userID,
		Status:        &published,
		OnlyPublicNow: true,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}, req.Latitude, req.Longitude)
}

func (s *PostService) listPosts(ctx context.Context, filter repository.PostListFilter, lat, lng float64) (*types.PostListResponse, error) {
	list, total, err := s.postRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return s.postListResponse(ctx, list, total, lat, lng)
}

func (s *PostService) postListResponse(ctx context.Context, list []*model.Post, total int64, lat, lng float64) (*types.PostListResponse, error) {
	items, err := s.buildPostListItems(ctx, list, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.PostListResponse{Total: total, List: items}, nil
}

func (s *PostService) canViewPost(ctx context.Context, userID uint, p *model.Post) bool {
	if userID > 0 && p.UserID > 0 && userID != p.UserID && s.userSvc != nil {
		if blocked, err := s.userSvc.IsBlockedEither(ctx, userID, p.UserID); err == nil && blocked {
			return false
		}
	}
	if p.Status == model.PostStatusPublished {
		return true
	}
	if userID == 0 || p.UserID == userID {
		return true
	}
	if p.Status != model.PostStatusDisabled {
		return false
	}
	signup, err := s.signupRepo.GetByPostAndUser(ctx, p.ID, userID)
	if err != nil {
		return false
	}
	return signup.Status == model.SignupStatusConfirmed || signup.Status == model.SignupStatusWaitlisted
}

func (s *PostService) Get(ctx context.Context, userID uint, id uint) (*types.PostDetail, error) {
	p, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("动态不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if !s.canViewPost(ctx, userID, p) {
		return nil, errs.NotFound("动态不存在")
	}
	d := &types.PostDetail{PostItem: s.toPostItem(p, s.postImageURLs(ctx, p))}
	if userID > 0 && p.UserID == userID {
		d.AuditRemark = ownerAuditRemark(p)
	}
	if p.UserID > 0 {
		publisher, err := s.loadPublisher(ctx, p.UserID)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		d.Publisher = publisher
	}
	extras, _ := s.extrasRepo.GetByPostID(ctx, p.ID)
	if extras != nil {
		d.Extras = extrasToPayload(extras)
		if extras.NeedSignup {
			counts, err := s.signupRepo.CountListByPost(ctx, p.ID)
			if err == nil {
				d.SignupSummary = &types.PostSignupSummary{
					Confirmed:    counts.Confirmed,
					Waitlisted:   counts.Waitlisted,
					Cancelled:    counts.Cancelled,
					NotCheckedIn: counts.NotCheckedIn,
					CheckedIn:    counts.CheckedIn,
				}
			}
		}
	}
	if userID > 0 {
		signup, err := s.signupRepo.GetByPostAndUser(ctx, p.ID, userID)
		if err == nil {
			d.MySignupStatus = signup.Status
			d.MyCheckedInAt = signup.CheckedInAt
		}
	}
	preview, err := s.discussPreview(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	d.DiscussPreview = preview
	return d, nil
}

func (s *PostService) discussPreview(ctx context.Context, postID uint) (*types.PostDiscussPreview, error) {
	if postID == 0 {
		return nil, nil
	}
	msg, err := s.msgRepo.LatestVisiblePost(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if msg == nil || strings.TrimSpace(msg.Content) == "" {
		return nil, nil
	}
	name := deactivatedDisplayName()
	if s.userSvc != nil && msg.FromUserID > 0 {
		profiles, err := s.userSvc.GetProfiles(ctx, []uint{msg.FromUserID})
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		if profile, ok := profiles[msg.FromUserID]; ok {
			name = utils.DisplayName(profile.Nickname)
		}
	}
	return &types.PostDiscussPreview{
		FromName:    name,
		ContentType: msg.ContentType,
		Content:     msg.Content,
	}, nil
}

func (s *PostService) buildPostListItems(ctx context.Context, list []*model.Post, userLat, userLng float64) ([]types.PostItem, error) {
	if len(list) == 0 {
		return []types.PostItem{}, nil
	}
	postIDs := make([]uint, len(list))
	for i, p := range list {
		postIDs[i] = p.ID
	}
	extrasList, err := s.extrasRepo.ListByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, fmt.Errorf("调用 ListByPostIDs 失败: %w", err)
	}
	extrasMap := make(map[uint]*model.PostExtra, len(extrasList))
	needSignupIDs := make([]uint, 0)
	for _, extras := range extrasList {
		extrasMap[extras.PostID] = extras
		if extras.NeedSignup {
			needSignupIDs = append(needSignupIDs, extras.PostID)
		}
	}
	signupMap := map[uint]repository.SignupSummary{}
	if len(needSignupIDs) > 0 {
		signupMap, err = s.signupRepo.CountSummaryByPostIDs(ctx, needSignupIDs)
		if err != nil {
			return nil, err
		}
	}

	publisherIDs := make([]uint, 0, len(list))
	seenPublisher := make(map[uint]struct{})
	imageIDs := make([][]uint, len(list))
	mediaIDs := make([]uint, 0)
	for i, p := range list {
		ids := parseMediaIDs(p.Images)
		imageIDs[i] = ids
		mediaIDs = append(mediaIDs, ids...)
		if p.UserID == 0 {
			continue
		}
		if _, ok := seenPublisher[p.UserID]; ok {
			continue
		}
		seenPublisher[p.UserID] = struct{}{}
		publisherIDs = append(publisherIDs, p.UserID)
	}
	profiles, err := s.userSvc.GetProfiles(ctx, publisherIDs)
	if err != nil {
		return nil, err
	}
	publisherMap := make(map[uint]*types.PostPublisher, len(publisherIDs))
	for id, user := range profiles {
		publisherMap[id] = &types.PostPublisher{
			ID:        user.ID,
			Nickname:  utils.DisplayName(user.Nickname),
			AvatarURL: user.AvatarURL,
		}
	}
	mediaURLs := map[uint]string{}
	if s.mediaSvc != nil && len(mediaIDs) > 0 {
		mediaURLs = s.mediaSvc.ResolveURLs(ctx, mediaIDs)
	}

	hasUserLocation := userLat != 0 || userLng != 0
	items := make([]types.PostItem, 0, len(list))
	for i, p := range list {
		item := s.toPostItem(p, mediaURLList(imageIDs[i], mediaURLs))
		if hasUserLocation && (p.Latitude != 0 || p.Longitude != 0) {
			item.Distance = utils.Haversine(userLat, userLng, p.Latitude, p.Longitude)
		}
		if extras, ok := extrasMap[p.ID]; ok {
			item.Extras = extrasToPayload(extras)
			if extras.NeedSignup {
				summary := signupMap[p.ID]
				item.SignupSummary = &types.PostSignupSummary{
					Confirmed:  summary.Confirmed,
					Waitlisted: summary.Waitlisted,
				}
			}
		}
		if pub, ok := publisherMap[p.UserID]; ok {
			item.Publisher = pub
		} else if p.UserID > 0 {
			item.Publisher = &types.PostPublisher{
				ID:       p.UserID,
				Nickname: deactivatedDisplayName(),
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *PostService) postImageURLs(ctx context.Context, p *model.Post) []string {
	if s.mediaSvc == nil {
		return []string{}
	}
	return s.mediaSvc.ResolveURLList(ctx, parseMediaIDs(p.Images))
}

func (s *PostService) toPostItem(p *model.Post, images []string) types.PostItem {
	if images == nil {
		images = []string{}
	}
	ids := parseMediaIDs(p.Images)
	return types.PostItem{
		ID:                  p.ID,
		RoomID:              p.RoomID,
		Category:            p.Category,
		Pinned:              p.Pinned,
		Title:               p.Title,
		Content:             p.Content,
		Images:              images,
		ImageIDs:            ids,
		PlaceName:           p.PlaceName,
		Latitude:            p.Latitude,
		Longitude:           p.Longitude,
		StartAt:             p.StartAt,
		EndAt:               p.EndAt,
		Status:              p.Status,
		PublisherID:         p.UserID,
		DiscussMessageCount: p.DiscussMessageCount,
		CreatedAt:           utils.FormatDateTime(p.CreatedAt),
		UpdatedAt:           utils.FormatDateTime(p.UpdatedAt),
	}
}

func mediaURLList(ids []uint, urls map[uint]string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if u := urls[id]; u != "" {
			out = append(out, u)
		}
	}
	return out
}

func (s *PostService) loadPublisher(ctx context.Context, userID uint) (*types.PostPublisher, error) {
	publisher := &types.PostPublisher{ID: userID}
	profiles, err := s.userSvc.GetProfiles(ctx, []uint{userID})
	if err != nil {
		return nil, err
	}
	user, ok := profiles[userID]
	if !ok {
		publisher.Nickname = deactivatedDisplayName()
		return publisher, nil
	}
	publisher.Nickname = utils.DisplayName(user.Nickname)
	publisher.AvatarURL = user.AvatarURL
	return publisher, nil
}

func extrasToPayload(extras *model.PostExtra) *types.PostExtrasPayload {
	if extras == nil || !postExtrasHasContent(extras) {
		return nil
	}
	return &types.PostExtrasPayload{
		Amount:           extras.Amount,
		Negotiable:       extras.Negotiable,
		NeedSignup:       extras.NeedSignup,
		NeedCheckIn:      extras.NeedSignup && extras.NeedCheckIn,
		MaxParticipants:  extras.MaxParticipants,
		AllowWaitlist:    extras.AllowWaitlist,
		SignupDeadlineAt: extras.SignupDeadlineAt,
		FeeType:          extras.FeeType,
		FeeDescription:   extras.FeeDescription,
	}
}
