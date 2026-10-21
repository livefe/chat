package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/logger"

	"github.com/redis/go-redis/v9"
)

const roomPresenceTTL = 90 * time.Second

type PresenceService struct {
	log        *logger.Logger
	cache      *cache.Cache
	userRepo   *UserStore
	instanceID string
}

func NewPresenceService(
	log *logger.Logger,
	c *cache.Cache,
	userRepo *UserStore,
	instanceID string,
) *PresenceService {
	return &PresenceService{
		log:        log,
		cache:      c,
		userRepo:   userRepo,
		instanceID: instanceID,
	}
}

func (s *PresenceService) MarkOnline(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(types.CacheWSPresenceKey, userID)
	now := strconv.FormatInt(time.Now().Unix(), 10)
	pipe := s.cache.Pipeline()
	pipe.HSet(ctx, key, s.instanceID, now)
	pipe.Expire(ctx, key, types.Expire45s*2)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *PresenceService) SaveOnlineLocation(ctx context.Context, userID uint, lat, lng float64) error {
	name := strconv.FormatUint(uint64(userID), 10)
	loc := &redis.GeoLocation{Name: name, Latitude: lat, Longitude: lng}
	return s.cache.GeoAdd(ctx, types.CacheWSOnlineGeoIndex, loc).Err()
}

func (s *PresenceService) RemoveOnlineGeo(ctx context.Context, userID uint) {
	if userID == 0 {
		return
	}
	name := strconv.FormatUint(uint64(userID), 10)
	if err := s.cache.ZRem(ctx, types.CacheWSOnlineGeoIndex, name).Err(); err != nil {
		s.log.Warn(ctx, "移除 online geo 失败", logger.Uint("user_id", userID), logger.Error(err))
	}
}

func (s *PresenceService) OnlineLocation(ctx context.Context, userID uint) (lat, lng float64, ok bool) {
	if userID == 0 {
		return 0, 0, false
	}
	name := strconv.FormatUint(uint64(userID), 10)
	pos, err := s.cache.GeoPos(ctx, types.CacheWSOnlineGeoIndex, name).Result()
	if err != nil || len(pos) == 0 || pos[0] == nil {
		return 0, 0, false
	}
	if pos[0].Latitude == 0 && pos[0].Longitude == 0 {
		return 0, 0, false
	}
	return pos[0].Latitude, pos[0].Longitude, true
}

func (s *PresenceService) EnsureOnlineGeo(ctx context.Context, userID uint, lat, lng float64) {
	if userID == 0 || (lat == 0 && lng == 0) {
		return
	}
	if err := s.SaveOnlineLocation(ctx, userID, lat, lng); err != nil {
		s.log.Warn(ctx, "EnsureOnlineGeo 写入失败", logger.Uint("user_id", userID), logger.Error(err))
		return
	}
	_ = s.MarkOnline(ctx, userID)
}

func (s *PresenceService) ClearOnline(ctx context.Context, userID uint) {
	key := fmt.Sprintf(types.CacheWSPresenceKey, userID)
	if err := s.cache.HDel(ctx, key, s.instanceID).Err(); err != nil {
		s.log.Warn(ctx, "清理 WebSocket presence 失败", logger.Uint("user_id", userID), logger.Error(err))
		return
	}
	if s.pruneUserPresence(ctx, key) {
		return
	}
	s.RemoveOnlineGeo(ctx, userID)
}

func (s *PresenceService) pruneUserPresence(ctx context.Context, key string) bool {
	all, err := s.cache.HGetAll(ctx, key).Result()
	if err != nil || len(all) == 0 {
		_ = s.cache.Del(ctx, key).Err()
		return false
	}
	now := time.Now().Unix()
	ttlSec := int64(types.Expire45s.Seconds())
	stale := make([]string, 0)
	alive := 0
	for field, val := range all {
		ts, err := strconv.ParseInt(val, 10, 64)
		if err != nil || now-ts > ttlSec {
			stale = append(stale, field)
			continue
		}
		alive++
	}
	if len(stale) > 0 {
		_ = s.cache.HDel(ctx, key, stale...).Err()
	}
	if alive == 0 {
		_ = s.cache.Del(ctx, key).Err()
		return false
	}
	return true
}

type presenceKind struct {
	keyFmt string
	label  string
	idKey  string
}

var (
	roomPresence = presenceKind{
		keyFmt: types.CacheRoomPresenceKey,
		label:  "聊天区",
		idKey:  "room_id",
	}
	postPresence = presenceKind{
		keyFmt: types.CachePostPresenceKey,
		label:  "动态",
		idKey:  "post_id",
	}
)

func (s *PresenceService) presenceField(userID uint) string {
	return strconv.FormatUint(uint64(userID), 10) + ":" + s.instanceID
}

func parsePresenceUserID(field string) (uint, bool) {
	base := field
	if i := strings.IndexByte(field, ':'); i >= 0 {
		base = field[:i]
	}
	id, err := strconv.ParseUint(base, 10, 64)
	return uint(id), err == nil
}

func (s *PresenceService) SetPresence(ctx context.Context, kind presenceKind, id, userID uint, present bool) bool {
	if id == 0 || userID == 0 {
		return false
	}
	key := fmt.Sprintf(kind.keyFmt, id)
	field := s.presenceField(userID)
	var err error
	if present {
		now := strconv.FormatInt(time.Now().Unix(), 10)
		err = s.cache.HSet(ctx, key, field, now).Err()
	} else {
		err = s.cache.HDel(ctx, key, field).Err()
	}
	if err != nil {
		s.log.Warn(ctx, "更新"+kind.label+"在线状态失败",
			logger.Uint(kind.idKey, id),
			logger.Uint("user_id", userID),
			logger.Error(err))
		return false
	}
	return true
}

func (s *PresenceService) TouchPresence(ctx context.Context, kind presenceKind, id, userID uint) (recreated bool) {
	if id == 0 || userID == 0 {
		return false
	}
	key := fmt.Sprintf(kind.keyFmt, id)
	field := s.presenceField(userID)
	now := strconv.FormatInt(time.Now().Unix(), 10)
	pipe := s.cache.Pipeline()
	existsCmd := pipe.HExists(ctx, key, field)
	setCmd := pipe.HSet(ctx, key, field, now)
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		s.log.Warn(ctx, "刷新"+kind.label+"在线状态失败",
			logger.Uint(kind.idKey, id),
			logger.Uint("user_id", userID),
			logger.Error(err))
		return false
	}
	exists, err := existsCmd.Result()
	if err != nil {
		return false
	}
	if err := setCmd.Err(); err != nil {
		s.log.Warn(ctx, "刷新"+kind.label+"在线状态失败",
			logger.Uint(kind.idKey, id),
			logger.Uint("user_id", userID),
			logger.Error(err))
		return false
	}
	return !exists
}

func aliveUserSet(all map[string]string, now, ttlSec int64) (map[uint]struct{}, []string) {
	aliveUsers := make(map[uint]struct{}, len(all))
	staleFields := make([]string, 0)
	for field, val := range all {
		ts, err := strconv.ParseInt(val, 10, 64)
		if err != nil || now-ts > ttlSec {
			staleFields = append(staleFields, field)
			continue
		}
		userID, ok := parsePresenceUserID(field)
		if !ok {
			staleFields = append(staleFields, field)
			continue
		}
		aliveUsers[userID] = struct{}{}
	}
	return aliveUsers, staleFields
}

func (s *PresenceService) OnlineCount(ctx context.Context, kind presenceKind, id uint) (int, error) {
	if id == 0 {
		return 0, nil
	}
	key := fmt.Sprintf(kind.keyFmt, id)
	all, err := s.cache.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	aliveUsers, staleFields := aliveUserSet(all, time.Now().Unix(), int64(roomPresenceTTL.Seconds()))
	if len(staleFields) > 0 {
		_ = s.cache.HDel(ctx, key, staleFields...).Err()
	}
	return len(aliveUsers), nil
}

func (s *PresenceService) RoomOnlineCount(ctx context.Context, roomID uint) (int, error) {
	return s.OnlineCount(ctx, roomPresence, roomID)
}

func (s *PresenceService) PostOnlineCount(ctx context.Context, postID uint) (int, error) {
	return s.OnlineCount(ctx, postPresence, postID)
}

func (s *PresenceService) ClearRoomPresence(ctx context.Context, roomID uint) {
	if roomID == 0 {
		return
	}
	key := fmt.Sprintf(types.CacheRoomPresenceKey, roomID)
	if err := s.cache.Del(ctx, key).Err(); err != nil {
		s.log.Warn(ctx, "清理聊天区 presence 失败", logger.Uint("room_id", roomID), logger.Error(err))
	}
}

func (s *PresenceService) ClearPostPresence(ctx context.Context, postID uint) {
	if postID == 0 {
		return
	}
	key := fmt.Sprintf(types.CachePostPresenceKey, postID)
	if err := s.cache.Del(ctx, key).Err(); err != nil {
		s.log.Warn(ctx, "清理帖子 presence 失败", logger.Uint("post_id", postID), logger.Error(err))
	}
}

const nearbyOnlineScanLimit = 2000

type nearbyUser struct {
	UserID   uint
	Lat      float64
	Lng      float64
	Distance float64
}

func (s *PresenceService) searchOnlineGeo(ctx context.Context, lat, lng float64, radiusMeters float64, limit int) ([]nearbyUser, error) {
	res, err := s.cache.GeoSearchLocation(ctx, types.CacheWSOnlineGeoIndex, &redis.GeoSearchLocationQuery{
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
	list := make([]nearbyUser, 0, len(res))
	for _, item := range res {
		id, err := strconv.ParseUint(item.Name, 10, 64)
		if err != nil {
			continue
		}
		list = append(list, nearbyUser{
			UserID:   uint(id),
			Lat:      item.Latitude,
			Lng:      item.Longitude,
			Distance: item.Dist,
		})
	}
	return list, nil
}

func (s *PresenceService) pruneStaleOnlineGeo(ctx context.Context, candidateIDs []uint, online map[uint]struct{}) {
	if len(candidateIDs) == 0 {
		return
	}
	pipe := s.cache.Pipeline()
	n := 0
	for _, id := range candidateIDs {
		if _, ok := online[id]; ok {
			continue
		}
		pipe.ZRem(ctx, types.CacheWSOnlineGeoIndex, strconv.FormatUint(uint64(id), 10))
		n++
	}
	if n == 0 {
		return
	}
	_, _ = pipe.Exec(ctx)
}

func (s *PresenceService) nearbyOnlineCandidates(ctx context.Context, lat, lng float64, radiusMeters float64, scanLimit int) ([]nearbyUser, error) {
	geoUsers, err := s.searchOnlineGeo(ctx, lat, lng, radiusMeters, scanLimit)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(geoUsers))
	for _, u := range geoUsers {
		ids = append(ids, u.UserID)
	}
	online := onlineUserIDSet(ctx, s.cache, ids)
	s.pruneStaleOnlineGeo(ctx, ids, online)

	out := make([]nearbyUser, 0, len(geoUsers))
	for _, u := range geoUsers {
		if _, ok := online[u.UserID]; ok {
			out = append(out, u)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Distance != out[j].Distance {
			return out[i].Distance < out[j].Distance
		}
		return out[i].UserID < out[j].UserID
	})
	return out, nil
}

func (s *PresenceService) NearbyOnlineUserCount(ctx context.Context, lat, lng float64, radiusMeters float64) (int, error) {
	users, err := s.nearbyOnlineCandidates(ctx, lat, lng, radiusMeters, nearbyOnlineScanLimit)
	if err != nil {
		return 0, err
	}
	ids := make([]uint, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.UserID)
	}
	if len(ids) == 0 {
		return 0, nil
	}
	profiles, err := s.userRepo.GetByIDsWithCache(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("调用 GetByIDsWithCache 失败: %w", err)
	}
	enabled := make(map[uint]struct{}, len(profiles))
	for _, p := range profiles {
		if p != nil && p.Status == model.UserStatusEnabled {
			enabled[p.ID] = struct{}{}
		}
	}
	count := 0
	for _, id := range ids {
		if _, ok := enabled[id]; ok {
			count++
		}
	}
	return count, nil
}

func isUserPresenceOnline(all map[string]string) bool {
	if len(all) == 0 {
		return false
	}
	now := time.Now().Unix()
	ttlSec := int64(types.Expire45s.Seconds())
	for _, val := range all {
		ts, err := strconv.ParseInt(val, 10, 64)
		if err == nil && now-ts <= ttlSec {
			return true
		}
	}
	return false
}

func onlineUserIDSet(ctx context.Context, c *cache.Cache, ids []uint) map[uint]struct{} {
	out := make(map[uint]struct{}, len(ids))
	if len(ids) == 0 {
		return out
	}
	pipe := c.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.HGetAll(ctx, fmt.Sprintf(types.CacheWSPresenceKey, id))
	}
	_, _ = pipe.Exec(ctx)
	for i, id := range ids {
		all, err := cmds[i].Result()
		if err != nil || len(all) == 0 {
			continue
		}
		if isUserPresenceOnline(all) {
			out[id] = struct{}{}
		}
	}
	return out
}
