package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"
)

func (s *RoomService) AdminGet(ctx context.Context, id uint) (*types.RoomDetail, error) {
	room, err := s.getRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.buildRoomDetail(room), nil
}

func (s *RoomService) AdminList(ctx context.Context, req *types.RoomListRequest) (*types.RoomListResponse, error) {
	filter := repository.RoomListFilter{
		ID: req.ID, Name: req.Name, Status: req.Status, Source: req.Source, PlaceType: req.PlaceType,
		Page: req.Page, PageSize: req.PageSize,
	}
	rooms, total, err := s.roomRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.RoomItem, 0, len(rooms))
	for _, r := range rooms {
		items = append(items, toRoomItem(r))
	}
	return &types.RoomListResponse{Total: total, List: items}, nil
}

func (s *RoomService) AdminCreate(ctx context.Context, employeeID uint, req *types.CreateRoomRequest) (uint, error) {
	return s.insertRoom(ctx, req, &model.Room{
		EmployeeID:      employeeID,
		Source:          model.RoomSourceAdmin,
		Status:          model.RoomStatusEnabled,
		RequireGeofence: req.RequireGeofence,
	}, true)
}

func (s *RoomService) AdminUpdate(ctx context.Context, req *types.UpdateRoomRequest) error {
	room, err := s.getRoom(ctx, req.ID)
	if err != nil {
		return err
	}
	if !isValidRoomPlaceType(req.PlaceType) {
		return errs.Invalid("请选择场所类型")
	}
	if err := assertClean(s.sensitive, req.Name, req.Description); err != nil {
		return err
	}
	room.Name = req.Name
	room.Description = req.Description
	room.PlaceType = req.PlaceType
	room.RequireGeofence = req.RequireGeofence
	if err := s.buildShapeFields(&req.CreateRoomRequest, room); err != nil {
		return err
	}
	if err := s.roomRepo.Update(ctx, room); err != nil {
		s.log.Error(ctx, "更新聊天区失败", logger.Error(err))
		return fmt.Errorf("更新失败: %w", err)
	}
	s.syncRoomGeoPresence(ctx, room)
	return nil
}

func (s *RoomService) AdminAudit(ctx context.Context, id uint, approve bool) error {
	room, err := s.getRoom(ctx, id)
	if err != nil {
		return err
	}
	if room.Status != model.RoomStatusPending {
		return errs.Forbidden("只有待审核状态可以审核")
	}
	if approve {
		room.Status = model.RoomStatusEnabled
	} else {
		room.Status = model.RoomStatusRejected
	}
	if err := s.roomRepo.Update(ctx, room); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	s.syncRoomGeoPresence(ctx, room)
	return nil
}

func (s *RoomService) AdminListNeighbors(ctx context.Context, req *types.AdminRoomNeighborsQuery) (*types.AdminRoomNeighborsResponse, error) {
	lat, lng := req.Lat, req.Lng
	excludeID := req.ExcludeID
	refName := ""
	if req.RoomID > 0 {
		room, err := s.getRoom(ctx, req.RoomID)
		if err != nil {
			return nil, err
		}
		lat, lng = room.CenterLat, room.CenterLng
		excludeID = room.ID
		refName = room.Name
	}
	if math.Abs(lat) < 1e-9 && math.Abs(lng) < 1e-9 {
		return &types.AdminRoomNeighborsResponse{List: []types.AdminRoomNeighborItem{}}, nil
	}
	radius := req.RadiusM
	limit := req.Limit
	statuses := []uint8{model.RoomStatusPending, model.RoomStatusEnabled}
	rooms, err := s.roomRepo.ListNearCenter(ctx, lat, lng, radius, excludeID, statuses, limit)
	if err != nil {
		s.log.Error(ctx, "查询邻近聊天区失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	type scored struct {
		item types.AdminRoomNeighborItem
		dist float64
	}
	scoredList := make([]scored, 0, len(rooms))
	for _, r := range rooms {
		if r.CenterLat == 0 && r.CenterLng == 0 {
			continue
		}
		dist := utils.Haversine(lat, lng, r.CenterLat, r.CenterLng)
		if dist > radius {
			continue
		}
		item := types.AdminRoomNeighborItem{
			ID:            r.ID,
			Name:          r.Name,
			PlaceType:     r.PlaceType,
			PlaceTypeText: utils.TextOf(roomPlaceTypeText, r.PlaceType),
			ShapeType:     r.ShapeType,
			CenterLat:     r.CenterLat,
			CenterLng:     r.CenterLng,
			Radius:        r.Radius,
			Status:        r.Status,
			StatusText:    utils.TextOf(roomStatusText, r.Status),
			Distance:      dist,
		}
		item.Suspect, item.SuspectReason = roomNeighborSuspect(refName, r.Name, dist)
		scoredList = append(scoredList, scored{item: item, dist: dist})
	}
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].item.Suspect != scoredList[j].item.Suspect {
			return scoredList[i].item.Suspect
		}
		return scoredList[i].dist < scoredList[j].dist
	})
	if len(scoredList) > req.Limit {
		scoredList = scoredList[:req.Limit]
	}
	out := make([]types.AdminRoomNeighborItem, 0, len(scoredList))
	for _, it := range scoredList {
		out = append(out, it.item)
	}
	return &types.AdminRoomNeighborsResponse{List: out}, nil
}

func (s *RoomService) AdminMap(ctx context.Context, req *types.AdminRoomMapQuery) (*types.AdminRoomMapResponse, error) {
	rooms, _, err := s.roomRepo.List(ctx, repository.RoomListFilter{
		Status:    req.Status,
		PlaceType: req.PlaceType,
		Source:    req.Source,
		Page:      1,
		PageSize:  req.Limit,
	})
	if err != nil {
		s.log.Error(ctx, "查询全景地图聊天区失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	out := make([]types.AdminRoomMapItem, 0, len(rooms))
	for _, r := range rooms {
		if r.CenterLat == 0 && r.CenterLng == 0 && r.ShapeType != model.RoomShapePolygon {
			continue
		}
		item := types.AdminRoomMapItem{
			ID:            r.ID,
			Name:          r.Name,
			PlaceType:     r.PlaceType,
			PlaceTypeText: utils.TextOf(roomPlaceTypeText, r.PlaceType),
			ShapeType:     r.ShapeType,
			CenterLat:     r.CenterLat,
			CenterLng:     r.CenterLng,
			Radius:        r.Radius,
			Status:        r.Status,
			StatusText:    utils.TextOf(roomStatusText, r.Status),
			Source:        r.Source,
			SourceText:    utils.TextOf(roomSourceText, r.Source),
		}
		if r.ShapeType == model.RoomShapePolygon {
			if pts, err := utils.DecodePolygon(r.Polygon); err == nil && len(pts) >= 3 {
				item.Polygon = pts
			}
		}
		out = append(out, item)
	}
	return &types.AdminRoomMapResponse{List: out}, nil
}

func (s *RoomService) AdminUpdateStatus(ctx context.Context, id uint, status uint8) error {
	room, err := s.getRoom(ctx, id)
	if err != nil {
		return err
	}
	room.Status = status
	if err := s.roomRepo.Update(ctx, room); err != nil {
		return fmt.Errorf("更新失败: %w", err)
	}
	s.syncRoomGeoPresence(ctx, room)
	if status != model.RoomStatusEnabled {
		s.closeRoomChannel(ctx, id, "聊天区已关闭")
	}
	return nil
}

func roomNeighborSuspect(refName, name string, dist float64) (bool, string) {
	if dist <= 300 {
		return true, "圆心过近"
	}
	if dist <= 800 && roomNameSimilar(refName, name) {
		return true, "名称相似且较近"
	}
	if roomNameSimilar(refName, name) && dist <= 1500 {
		return true, "名称相似"
	}
	return false, ""
}

func roomNameSimilar(a, b string) bool {
	na := normalizeRoomName(a)
	nb := normalizeRoomName(b)
	if na == "" || nb == "" {
		return false
	}
	if na == nb {
		return true
	}
	return strings.Contains(na, nb) || strings.Contains(nb, na)
}

func normalizeRoomName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	replacer := strings.NewReplacer(
		" ", "", "　", "", "-", "", "_", "",
		"聊天区", "", "小区", "", "社区", "",
	)
	return replacer.Replace(s)
}
