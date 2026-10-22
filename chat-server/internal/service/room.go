package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/utils"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type roomChannelNotifier interface {
	CloseRoom(ctx context.Context, roomID uint, reason string)
}

type roomMessageReader interface {
	ListNearby(ctx context.Context, lat, lng float64, radiusMeters float64, beforeID uint, limit int) ([]types.MessageItem, error)
	ListRoom(ctx context.Context, userID uint, roomID uint, lat, lng float64, beforeID uint, limit int) ([]types.MessageItem, error)
}

type RoomService struct {
	log          *logger.Logger
	cache        *cache.Cache
	roomRepo     repository.RoomRepository
	employeeRepo repository.EmployeeRepository
	userSvc      *UserService
	presence     *PresenceService
	sensitive    *SensitiveWordService
	channels     roomChannelNotifier
	messages     roomMessageReader
}

func (s *RoomService) SetMessageReader(r roomMessageReader) {
	s.messages = r
}

var (
	roomStatusText = map[uint8]string{
		model.RoomStatusPending:  "待审核",
		model.RoomStatusEnabled:  "正常",
		model.RoomStatusRejected: "驳回",
		model.RoomStatusDisabled: "禁用",
	}
	roomSourceText = map[uint8]string{
		model.RoomSourceAdmin: "后台",
		model.RoomSourceUser:  "用户",
	}
	roomPlaceTypeText = map[uint8]string{
		model.RoomPlaceCommunity: "小区",
		model.RoomPlaceSchool:    "学校",
		model.RoomPlaceBusiness:  "商圈",
		model.RoomPlaceScenic:    "景点",
		model.RoomPlaceCampus:    "园区",
		model.RoomPlaceTransit:   "交通枢纽",
		model.RoomPlaceOther:     "其他",
	}
)

func isValidRoomPlaceType(v uint8) bool {
	_, ok := roomPlaceTypeText[v]
	return ok
}

func NewRoomService(
	log *logger.Logger,
	c *cache.Cache,
	roomRepo repository.RoomRepository,
	employeeRepo repository.EmployeeRepository,
	userSvc *UserService,
	presence *PresenceService,
	sensitive *SensitiveWordService,
	channels roomChannelNotifier,
) *RoomService {
	return &RoomService{
		log:          log,
		cache:        c,
		roomRepo:     roomRepo,
		employeeRepo: employeeRepo,
		userSvc:      userSvc,
		presence:     presence,
		sensitive:    sensitive,
		channels:     channels,
	}
}

func (s *RoomService) CountByStatus(ctx context.Context, status uint8) (int64, error) {
	return s.roomRepo.CountByStatus(ctx, status)
}

func (s *RoomService) closeRoomChannel(ctx context.Context, roomID uint, reason string) {
	if roomID == 0 {
		return
	}
	s.channels.CloseRoom(ctx, roomID, reason)
}

func (s *RoomService) Create(ctx context.Context, userID uint, req *types.CreateRoomRequest) (uint, error) {
	if req.ShapeType != model.RoomShapeCircle {
		return 0, errs.Invalid("普通用户仅能创建圆形聊天区")
	}
	return s.insertRoom(ctx, req, &model.Room{
		UserID:          userID,
		Source:          model.RoomSourceUser,
		Status:          model.RoomStatusPending,
		RequireGeofence: req.RequireGeofence,
	}, false)
}

func (s *RoomService) ListMyApplications(ctx context.Context, userID uint, req *types.MyRoomApplicationsQuery) (*types.RoomListResponse, error) {
	source := model.RoomSourceUser
	list, total, err := s.roomRepo.List(ctx, repository.RoomListFilter{
		UserID:   &userID,
		Source:   &source,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		s.log.Error(ctx, "查询聊天区申请失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	items := make([]types.RoomItem, 0, len(list))
	for _, r := range list {
		it := toRoomItem(r)
		it.StatusText = roomApplicationStatusText(r.Status)
		items = append(items, it)
	}
	return &types.RoomListResponse{Total: total, List: items}, nil
}

func roomApplicationStatusText(status uint8) string {
	switch status {
	case model.RoomStatusPending:
		return "审核中"
	case model.RoomStatusEnabled:
		return "已通过"
	case model.RoomStatusRejected:
		return "已驳回"
	case model.RoomStatusDisabled:
		return "已停用"
	default:
		return "未知"
	}
}

func (s *RoomService) insertRoom(ctx context.Context, req *types.CreateRoomRequest, room *model.Room, syncGeo bool) (uint, error) {
	if !isValidRoomPlaceType(req.PlaceType) {
		return 0, errs.Invalid("请选择场所类型")
	}
	if err := assertClean(s.sensitive, req.Name, req.Description); err != nil {
		return 0, err
	}
	room.Name = req.Name
	room.Description = req.Description
	room.PlaceType = req.PlaceType
	if err := s.buildShapeFields(req, room); err != nil {
		return 0, err
	}
	if err := s.roomRepo.Create(ctx, room); err != nil {
		s.log.Error(ctx, "创建聊天区失败", logger.Error(err))
		return 0, fmt.Errorf("创建失败: %w", err)
	}
	if syncGeo {
		s.syncRoomGeoPresence(ctx, room)
	}
	return room.ID, nil
}

func toRoomItem(r *model.Room) types.RoomItem {
	item := types.RoomItem{
		ID:              r.ID,
		Name:            r.Name,
		Description:     r.Description,
		PlaceType:       r.PlaceType,
		PlaceTypeText:   utils.TextOf(roomPlaceTypeText, r.PlaceType),
		ShapeType:       r.ShapeType,
		CenterLat:       r.CenterLat,
		CenterLng:       r.CenterLng,
		Radius:          r.Radius,
		UserID:          r.UserID,
		EmployeeID:      r.EmployeeID,
		Source:          r.Source,
		SourceText:      utils.TextOf(roomSourceText, r.Source),
		Status:          r.Status,
		StatusText:      utils.TextOf(roomStatusText, r.Status),
		RequireGeofence: r.RequireGeofence,
		CreatedAt:       utils.FormatDateTime(r.CreatedAt),
		UpdatedAt:       utils.FormatDateTime(r.UpdatedAt),
	}
	if r.Status == model.RoomStatusRejected {
		item.AuditRemark = r.AuditRemark
	}
	return item
}

func (s *RoomService) Get(ctx context.Context, id uint) (*types.RoomDetail, error) {
	room, err := s.getRoom(ctx, id)
	if err != nil {
		return nil, err
	}
	if room.Status != model.RoomStatusEnabled {
		return nil, errs.Forbidden("聊天区不可用")
	}
	return s.buildRoomDetail(room), nil
}

func (s *RoomService) getRoom(ctx context.Context, id uint) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("聊天区不存在")
		}
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	return room, nil
}

func (s *RoomService) buildRoomDetail(room *model.Room) *types.RoomDetail {
	d := &types.RoomDetail{RoomItem: toRoomItem(room)}
	if room.ShapeType == model.RoomShapePolygon {
		if pts, err := utils.DecodePolygon(room.Polygon); err == nil {
			d.Polygon = pts
		}
	}
	return d
}

func (s *RoomService) ListNearby(ctx context.Context, lat, lng float64, radiusMeters float64, limit int) ([]*types.NearbyRoomItem, error) {
	geos, err := s.searchNearbyRoomGeo(ctx, lat, lng, radiusMeters, limit)
	if err != nil {
		s.log.Error(ctx, "查询附近聊天区失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	if len(geos) == 0 {
		return []*types.NearbyRoomItem{}, nil
	}
	ids := make([]uint, 0, len(geos))
	for _, g := range geos {
		ids = append(ids, g.RoomID)
	}
	rooms, err := s.roomRepo.GetByIDs(ctx, ids)
	if err != nil {
		s.log.Error(ctx, "查询附近聊天区失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	byID := make(map[uint]*model.Room, len(rooms))
	for _, r := range rooms {
		if r.Status == model.RoomStatusEnabled {
			byID[r.ID] = r
		}
	}
	out := make([]*types.NearbyRoomItem, 0, len(geos))
	for _, g := range geos {
		r, ok := byID[g.RoomID]
		if !ok {
			continue
		}
		out = append(out, &types.NearbyRoomItem{
			ID:              r.ID,
			Name:            r.Name,
			PlaceType:       r.PlaceType,
			PlaceTypeText:   utils.TextOf(roomPlaceTypeText, r.PlaceType),
			ShapeType:       r.ShapeType,
			CenterLat:       r.CenterLat,
			CenterLng:       r.CenterLng,
			Radius:          r.Radius,
			Distance:        g.Distance,
			RequireGeofence: r.RequireGeofence,
		})
	}
	return out, nil
}

func (s *RoomService) searchNearbyRoomGeo(ctx context.Context, lat, lng float64, radiusMeters float64, limit int) ([]roomGeoHit, error) {
	res, err := s.cache.GeoSearchLocation(ctx, types.CacheRoomGeoIndex, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radiusMeters,
			RadiusUnit: "m",
			Sort:       "ASC",
			Count:      limit,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil {
		return nil, err
	}
	list := make([]roomGeoHit, 0, len(res))
	for _, item := range res {
		id, err := strconv.ParseUint(item.Name, 10, 64)
		if err != nil {
			continue
		}
		list = append(list, roomGeoHit{
			RoomID:   uint(id),
			Lat:      item.Latitude,
			Lng:      item.Longitude,
			Distance: item.Dist,
		})
	}
	return list, nil
}

func (s *RoomService) NearbyBootstrap(ctx context.Context, userID uint, lat, lng float64, radiusMeters float64, limit int, withMessages bool) (*types.NearbyBootstrapResponse, error) {
	resp, err := s.nearbySnapshot(ctx, userID, lat, lng, radiusMeters, limit)
	if err != nil || !withMessages {
		return resp, err
	}
	msgs, err := s.messages.ListNearby(ctx, lat, lng, radiusMeters, 0, limit)
	if err != nil {
		return nil, err
	}
	resp.Messages = msgs
	return resp, nil
}

func (s *RoomService) RoomBootstrap(ctx context.Context, userID, roomID uint, lat, lng float64, limit int, withMessages bool) (*types.RoomBootstrapResponse, error) {
	resp, err := s.roomSnapshot(ctx, roomID)
	if err != nil || !withMessages {
		return resp, err
	}
	msgs, err := s.messages.ListRoom(ctx, userID, roomID, lat, lng, 0, limit)
	if err != nil {
		return nil, err
	}
	resp.Messages = msgs
	return resp, nil
}

func (s *RoomService) nearbySnapshot(ctx context.Context, userID uint, lat, lng float64, radiusMeters float64, limit int) (*types.NearbyBootstrapResponse, error) {
	if userID > 0 && !(lat == 0 && lng == 0) {
		s.presence.EnsureOnlineGeo(ctx, userID, lat, lng)
	}
	rooms, err := s.ListNearby(ctx, lat, lng, radiusMeters, limit)
	if err != nil {
		return nil, err
	}
	count, err := s.presence.NearbyOnlineUserCount(ctx, lat, lng, radiusMeters)
	if err != nil {
		s.log.Error(ctx, "查询附近在线人数失败", logger.Error(err))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	resp := &types.NearbyBootstrapResponse{
		OnlineCount: count,
		Rooms:       make([]types.NearbyRoomItem, 0, len(rooms)),
	}
	for _, r := range rooms {
		resp.Rooms = append(resp.Rooms, *r)
	}
	return resp, nil
}

func (s *RoomService) roomSnapshot(ctx context.Context, roomID uint) (*types.RoomBootstrapResponse, error) {
	detail, err := s.Get(ctx, roomID)
	if err != nil {
		return nil, err
	}
	online, err := s.presence.RoomOnlineCount(ctx, roomID)
	if err != nil {
		s.log.Error(ctx, "查询聊天区在线人数失败", logger.Error(err), logger.Uint("roomId", roomID))
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	return &types.RoomBootstrapResponse{
		Room:        *detail,
		OnlineCount: online,
	}, nil
}

func (s *RoomService) buildShapeFields(req *types.CreateRoomRequest, room *model.Room) error {
	switch req.ShapeType {
	case model.RoomShapeCircle:
		if req.CenterLat == 0 && req.CenterLng == 0 {
			return errs.Invalid("圆形聊天区缺少中心坐标")
		}
		if req.Radius == 0 {
			return errs.Invalid("圆形聊天区半径不能为0")
		}
		room.ShapeType = model.RoomShapeCircle
		room.CenterLat = req.CenterLat
		room.CenterLng = req.CenterLng
		room.Radius = req.Radius
		room.Polygon = nil
	case model.RoomShapePolygon:
		if len(req.Polygon) < 3 {
			return errs.Invalid("多边形至少需要3个顶点")
		}
		pts := make([][2]float64, 0, len(req.Polygon))
		for _, p := range req.Polygon {
			if len(p) != 2 {
				return errs.Invalid("多边形顶点格式错误")
			}
			pts = append(pts, [2]float64{p[0], p[1]})
		}
		lat, lng := utils.Centroid(pts)
		radius := utils.BoundingRadius(lat, lng, pts)
		encoded, err := utils.EncodePolygon(pts)
		if err != nil {
			return fmt.Errorf("多边形编码失败: %w", err)
		}
		room.ShapeType = model.RoomShapePolygon
		room.CenterLat = lat
		room.CenterLng = lng
		room.Radius = radius
		room.Polygon = encoded
	default:
		return errs.Invalid("不支持的形状类型")
	}
	return nil
}

func (s *RoomService) syncRoomGeoPresence(ctx context.Context, room *model.Room) {
	if room == nil {
		return
	}
	if room.Status == model.RoomStatusEnabled {
		if err := s.upsertRoomGeo(ctx, room.ID, room.CenterLat, room.CenterLng); err != nil {
			s.log.Warn(ctx, "同步房间到 Redis GEO 失败", logger.Error(err))
		}
		return
	}
	_ = s.removeRoomGeo(ctx, room.ID)
}

func (s *RoomService) upsertRoomGeo(ctx context.Context, roomID uint, lat, lng float64) error {
	return s.cache.GeoAdd(ctx, types.CacheRoomGeoIndex, &redis.GeoLocation{
		Name:      strconv.FormatUint(uint64(roomID), 10),
		Latitude:  lat,
		Longitude: lng,
	}).Err()
}

func (s *RoomService) removeRoomGeo(ctx context.Context, roomID uint) error {
	return s.cache.ZRem(ctx, types.CacheRoomGeoIndex, strconv.FormatUint(uint64(roomID), 10)).Err()
}

func (s *RoomService) IsUserInRoomGeofence(room *model.Room, lat, lng float64) bool {
	if room == nil || !room.RequireGeofence {
		return true
	}
	if lat == 0 && lng == 0 {
		return false
	}
	switch room.ShapeType {
	case model.RoomShapeCircle:
		return utils.InCircle(lat, lng, room.CenterLat, room.CenterLng, float64(room.Radius))
	case model.RoomShapePolygon:
		pts, err := utils.DecodePolygon(room.Polygon)
		if err != nil || len(pts) < 3 {
			return false
		}
		return utils.InPolygon(lat, lng, pts)
	default:
		return false
	}
}

func (s *RoomService) ResolveUserLocation(ctx context.Context, userID uint, lat, lng float64) (float64, float64) {
	if lat != 0 || lng != 0 {
		return lat, lng
	}
	if fbLat, fbLng, ok := s.presence.OnlineLocation(ctx, userID); ok {
		return fbLat, fbLng
	}
	return 0, 0
}

func roomEnabled(room *model.Room) bool {
	return room != nil && room.Status == model.RoomStatusEnabled
}

func (s *RoomService) roomAccessErr(room *model.Room, lat, lng float64) error {
	if !roomEnabled(room) {
		return errs.Forbidden("聊天区不可用")
	}
	if !s.IsUserInRoomGeofence(room, lat, lng) {
		return errs.Forbidden("您不在该聊天区范围内，无法聊天")
	}
	return nil
}

func (s *RoomService) CheckAccess(ctx context.Context, roomID uint, lat, lng float64) error {
	room, err := s.getRoom(ctx, roomID)
	if err != nil {
		return err
	}
	return s.roomAccessErr(room, lat, lng)
}

func (s *RoomService) CheckAccessBatch(ctx context.Context, roomIDs []uint, lat, lng float64) (map[uint]error, error) {
	if len(roomIDs) == 0 {
		return nil, nil
	}
	rooms, err := s.roomRepo.GetByIDs(ctx, roomIDs)
	if err != nil {
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	byID := make(map[uint]*model.Room, len(rooms))
	for _, r := range rooms {
		byID[r.ID] = r
	}
	out := make(map[uint]error)
	for _, id := range roomIDs {
		room, ok := byID[id]
		if !ok {
			out[id] = errs.NotFound("聊天区不存在")
			continue
		}
		if err := s.roomAccessErr(room, lat, lng); err != nil {
			out[id] = err
		}
	}
	return out, nil
}

type roomGeoHit struct {
	RoomID   uint
	Lat      float64
	Lng      float64
	Distance float64
}

func (s *RoomService) GetRoomsByIDs(ctx context.Context, ids []uint) ([]*model.Room, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	return s.roomRepo.GetByIDs(ctx, ids)
}
