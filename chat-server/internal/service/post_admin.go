package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/utils"

	"gorm.io/gorm"
)

const reportDisableRemark = "因举报被下架"

func (s *PostService) AdminList(ctx context.Context, req *types.AdminPostListRequest) (*types.AdminPostListResponse, error) {
	list, total, err := s.postRepo.List(ctx, repository.PostListFilter{
		ID:               req.ID,
		Status:           req.Status,
		UserID:           req.PublisherID,
		PublisherKeyword: strings.TrimSpace(req.PublisherKeyword),
		RoomID:           req.RoomID,
		Category:         req.Category,
		Keyword:          strings.TrimSpace(req.Keyword),
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		SortAsc:          req.Sort == "asc",
		IncludeDeleted:   req.IncludeDeleted,
		Page:             req.Page,
		PageSize:         req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.buildAdminPostItems(ctx, list)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.AdminPostListResponse{Total: total, List: items}, nil
}

func (s *PostService) AdminGet(ctx context.Context, id uint) (*types.AdminPostItem, error) {
	p, err := s.postRepo.GetByIDUnscoped(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("动态不存在")
		}
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.buildAdminPostItems(ctx, []*model.Post{p})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	item := items[0]
	if item.Extras != nil && item.Extras.NeedSignup {
		counts, err := s.signupRepo.CountListByPost(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		item.SignupSummary = &types.PostSignupSummary{
			Confirmed:    counts.Confirmed,
			Waitlisted:   counts.Waitlisted,
			Cancelled:    counts.Cancelled,
			NotCheckedIn: counts.NotCheckedIn,
			CheckedIn:    counts.CheckedIn,
		}
	}
	return &item, nil
}

func (s *PostService) buildAdminPostItems(ctx context.Context, list []*model.Post) ([]types.AdminPostItem, error) {
	base, err := s.buildPostListItems(ctx, list, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return []types.AdminPostItem{}, nil
	}

	userIDs := make([]uint, 0, len(list))
	roomIDs := make([]uint, 0, len(list))
	employeeIDs := make([]uint, 0, len(list))
	for _, p := range list {
		if p.UserID > 0 {
			userIDs = append(userIDs, p.UserID)
		}
		if p.RoomID > 0 {
			roomIDs = append(roomIDs, p.RoomID)
		}
		if p.AuditedBy > 0 {
			employeeIDs = append(employeeIDs, p.AuditedBy)
		}
	}
	profiles, err := s.userSvc.GetProfiles(ctx, uniqueUints(userIDs))
	if err != nil {
		return nil, err
	}
	roomNames := make(map[uint]string, len(roomIDs))
	if len(roomIDs) > 0 {
		rooms, err := s.roomSvc.GetRoomsByIDs(ctx, uniqueUints(roomIDs))
		if err != nil {
			return nil, err
		}
		for _, r := range rooms {
			roomNames[r.ID] = r.Name
		}
	}
	employeeNames, err := employeeNameMap(ctx, s.employeeRepo, employeeIDs)
	if err != nil {
		return nil, err
	}

	out := make([]types.AdminPostItem, len(list))
	for i, p := range list {
		item := types.AdminPostItem{
			PostItem:      base[i],
			RoomName:      roomNames[p.RoomID],
			Deleted:       p.DeletedAt.Valid,
			AuditedByName: employeeNames[p.AuditedBy],
			AuditedAt:     formatSignupUnix(p.AuditedAt),
		}
		item.AuditRemark = p.AuditRemark
		if profile, ok := profiles[p.UserID]; ok {
			item.PublisherMobile = utils.DesensitizeMobile(profile.Mobile)
		}
		out[i] = item
	}
	return out, nil
}

func employeeNameMap(ctx context.Context, repo repository.EmployeeRepository, ids []uint) (map[uint]string, error) {
	out := make(map[uint]string)
	ids = uniqueUints(ids)
	if len(ids) == 0 {
		return out, nil
	}
	employees, err := repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, e := range employees {
		name := e.Name
		if name == "" {
			name = utils.DesensitizeMobile(e.Mobile)
		}
		out[e.ID] = name
	}
	return out, nil
}

func (s *PostService) AdminMap(ctx context.Context, req *types.AdminPostMapQuery) (*types.AdminPostMapResponse, error) {
	counts, err := s.postRepo.CountGroupedByRoom(ctx, req.Status, req.Category)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	roomIDs := make([]uint, 0, len(counts))
	countMap := make(map[uint]int64, len(counts))
	for _, row := range counts {
		if row.RoomID == 0 {
			continue
		}
		roomIDs = append(roomIDs, row.RoomID)
		countMap[row.RoomID] = row.Count
	}

	roomsOut := make([]types.AdminPostMapRoomItem, 0, len(roomIDs))
	if len(roomIDs) > 0 {
		rooms, err := s.roomSvc.GetRoomsByIDs(ctx, roomIDs)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		for _, r := range rooms {
			if r.CenterLat == 0 && r.CenterLng == 0 && r.ShapeType != model.RoomShapePolygon {
				continue
			}
			item := types.AdminPostMapRoomItem{
				RoomID:         r.ID,
				Name:           r.Name,
				PlaceType:      r.PlaceType,
				PlaceTypeText:  utils.TextOf(roomPlaceTypeText, r.PlaceType),
				ShapeType:      r.ShapeType,
				CenterLat:      r.CenterLat,
				CenterLng:      r.CenterLng,
				Radius:         r.Radius,
				RoomStatus:     r.Status,
				RoomStatusText: utils.TextOf(roomStatusText, r.Status),
				PostCount:      countMap[r.ID],
			}
			if r.ShapeType == model.RoomShapePolygon {
				if pts, err := utils.DecodePolygon(r.Polygon); err == nil && len(pts) >= 3 {
					item.Polygon = pts
				}
			}
			roomsOut = append(roomsOut, item)
		}
	}

	indep, err := s.postRepo.ListIndependentForMap(ctx, req.Status, req.Category, req.IndependentLimit)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	indepOut := make([]types.AdminPostMapIndependentItem, 0, len(indep))
	for _, p := range indep {
		indepOut = append(indepOut, types.AdminPostMapIndependentItem{
			ID:        p.ID,
			Title:     p.Title,
			Category:  p.Category,
			PlaceName: p.PlaceName,
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
			Status:    p.Status,
			CreatedAt: utils.FormatDateTime(p.CreatedAt),
		})
	}

	return &types.AdminPostMapResponse{
		Rooms:       roomsOut,
		Independent: indepOut,
	}, nil
}

func postStatusTransitionAllowed(from, to uint8) bool {
	switch to {
	case model.PostStatusPublished:
		return from == model.PostStatusPending || from == model.PostStatusRejected || from == model.PostStatusDisabled
	case model.PostStatusRejected:
		return from == model.PostStatusPending
	case model.PostStatusDisabled:
		return from == model.PostStatusPublished
	default:
		return false
	}
}

func validatePostStatusReason(status uint8, reason string) error {
	if reason != "" {
		return nil
	}
	switch status {
	case model.PostStatusRejected:
		return errs.Invalid("请填写驳回原因")
	case model.PostStatusDisabled:
		return errs.Invalid("请填写下架原因")
	}
	return nil
}

func (s *PostService) applyAdminStatus(ctx context.Context, p *model.Post, employeeID uint, status uint8, reason string) error {
	p.Status = status
	p.AuditRemark = reason
	p.AuditedBy = employeeID
	p.AuditedAt = time.Now().Unix()
	if err := s.postRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	if status != model.PostStatusPublished {
		closeReason := "动态讨论已结束"
		if status == model.PostStatusDisabled {
			closeReason = "动态已被下架"
		}
		s.closePostChannel(ctx, p.ID, closeReason)
	}
	return nil
}

func (s *PostService) AdminUpdateStatus(ctx context.Context, employeeID uint, req *types.UpdatePostStatusRequest) error {
	reason := strings.TrimSpace(req.Reason)
	if err := validatePostStatusReason(req.Status, reason); err != nil {
		return err
	}
	p, err := s.postRepo.GetByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if !postStatusTransitionAllowed(p.Status, req.Status) {
		return errs.Conflict(fmt.Sprintf("动态当前为「%s」，不能改为「%s」",
			utils.TextOf(postStatusText, p.Status), utils.TextOf(postStatusText, req.Status)))
	}
	return s.applyAdminStatus(ctx, p, employeeID, req.Status, reason)
}

func (s *PostService) AdminBatchUpdateStatus(ctx context.Context, employeeID uint, req *types.BatchUpdatePostStatusRequest) (*types.ContentBatchResult, error) {
	reason := strings.TrimSpace(req.Reason)
	if err := validatePostStatusReason(req.Status, reason); err != nil {
		return nil, err
	}
	ids := uniqueUints(req.IDs)
	posts, err := s.postRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	result := &types.ContentBatchResult{Skipped: len(ids) - len(posts)}
	for _, p := range posts {
		if !postStatusTransitionAllowed(p.Status, req.Status) {
			result.Skipped++
			continue
		}
		if err := s.applyAdminStatus(ctx, p, employeeID, req.Status, reason); err != nil {
			return nil, err
		}
		result.Updated++
	}
	return result, nil
}

func (s *PostService) DisableByReport(ctx context.Context, postID uint) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if p.Status == model.PostStatusDisabled {
		return nil
	}
	return s.applyAdminStatus(ctx, p, 0, model.PostStatusDisabled, reportDisableRemark)
}

func (s *PostService) EnsurePostMessageAccess(ctx context.Context, userID uint, postID uint) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	if userID > 0 && p.UserID > 0 && userID != p.UserID && s.userSvc != nil {
		blocked, err := s.userSvc.IsBlockedEither(ctx, userID, p.UserID)
		if err != nil {
			return fmt.Errorf("查询失败: %w", err)
		}
		if blocked {
			return errs.NotFound("动态不存在")
		}
	}
	if !postPublished(p) {
		if p.Status == model.PostStatusDisabled {
			return errs.Forbidden("动态已下架")
		}
		return errs.Forbidden("动态未发布")
	}
	now := time.Now().Unix()
	if p.EndAt > 0 && now > p.EndAt {
		return errs.Forbidden("动态已过期")
	}
	return nil
}
