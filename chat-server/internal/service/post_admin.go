package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/utils"

	"gorm.io/gorm"
)

func (s *PostService) AdminList(ctx context.Context, req *types.AdminPostListRequest) (*types.PostListResponse, error) {
	list, total, err := s.postRepo.List(ctx, repository.PostListFilter{
		Status:   req.Status,
		UserID:   req.UserID,
		RoomID:   req.RoomID,
		Category: req.Category,
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items, err := s.buildPostListItems(ctx, list, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.PostListResponse{Total: total, List: items}, nil
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

func (s *PostService) AdminGet(ctx context.Context, id uint) (*types.PostDetail, error) {
	return s.Get(ctx, 0, id)
}

func (s *PostService) AdminUpdateStatus(ctx context.Context, id uint, status uint8) error {
	p, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NotFound("动态不存在")
		}
		return fmt.Errorf("查询失败: %w", err)
	}
	p.Status = status
	if err := s.postRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	if status != model.PostStatusPublished {
		reason := "动态讨论已结束"
		if status == model.PostStatusDisabled {
			reason = "动态已被下架"
		}
		s.closePostChannel(ctx, id, reason)
	}
	return nil
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
