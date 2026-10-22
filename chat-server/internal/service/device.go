package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/errs"
	"app/pkg/logger"
	"app/pkg/push"
	"app/pkg/utils"
)

type DeviceService struct {
	log        *logger.Logger
	cache      *cache.Cache
	deviceRepo repository.DeviceRepository
	dispatch   *push.Dispatcher
}

func NewDeviceService(
	log *logger.Logger,
	c *cache.Cache,
	deviceRepo repository.DeviceRepository,
	dispatch *push.Dispatcher,
) *DeviceService {
	if dispatch == nil {
		dispatch = push.BuildDispatchers(push.XiaomiConfig{})
	}
	return &DeviceService{
		log:        log,
		cache:      c,
		deviceRepo: deviceRepo,
		dispatch:   dispatch,
	}
}

func (s *DeviceService) Register(ctx context.Context, userID uint, req *types.DeviceRegisterCommand) error {
	if userID == 0 || req == nil {
		return errs.Invalid("参数错误")
	}
	d := &model.Device{
		UserID:       userID,
		DeviceID:     strings.TrimSpace(req.DeviceID),
		Platform:     req.Platform,
		Vendor:       req.Vendor,
		PushToken:    strings.TrimSpace(req.PushToken),
		AppVersion:   req.AppVersion,
		OSVersion:    req.OSVersion,
		PushEnabled:  true,
		LastActiveAt: time.Now().Unix(),
	}
	if err := s.deviceRepo.Upsert(ctx, d); err != nil {
		s.log.Error(ctx, "绑定推送设备失败", logger.Error(err))
		return fmt.Errorf("绑定推送设备失败: %w", err)
	}
	s.log.Info(ctx, "推送设备已绑定",
		logger.Uint("user_id", userID),
		logger.String("vendor", d.Vendor),
		logger.String("device_id", d.DeviceID))
	return nil
}

func (s *DeviceService) Unbind(ctx context.Context, userID uint, deviceID string) error {
	if userID == 0 {
		return errs.Invalid("参数错误")
	}
	if err := s.deviceRepo.DeleteByUserAndDevice(ctx, userID, deviceID); err != nil {
		s.log.Error(ctx, "解绑推送设备失败", logger.Error(err))
		return fmt.Errorf("解绑推送设备失败: %w", err)
	}
	return nil
}

func (s *DeviceService) DeleteByUserID(ctx context.Context, userID uint) error {
	if userID == 0 {
		return nil
	}
	return s.deviceRepo.DeleteByUserID(ctx, userID)
}

func (s *DeviceService) AdminListByUser(ctx context.Context, req *types.AdminUserSubListRequest) (*types.AdminDeviceListResponse, error) {
	rows, total, err := s.deviceRepo.ListByUserID(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		s.log.Error(ctx, "查询用户设备失败", logger.Uint("user_id", req.UserID), logger.Error(err))
		return nil, fmt.Errorf("查询用户设备失败: %w", err)
	}
	list := make([]types.AdminDeviceItem, 0, len(rows))
	for _, d := range rows {
		list = append(list, types.AdminDeviceItem{
			ID:           d.ID,
			DeviceID:     d.DeviceID,
			Platform:     d.Platform,
			Vendor:       d.Vendor,
			AppVersion:   d.AppVersion,
			OSVersion:    d.OSVersion,
			PushEnabled:  d.PushEnabled,
			LastActiveAt: d.LastActiveAt,
			CreatedAt:    utils.FormatDateTime(d.CreatedAt),
		})
	}
	return &types.AdminDeviceListResponse{Total: total, List: list}, nil
}

func (s *DeviceService) NotifyPrivateOffline(ctx context.Context, toUserID uint, item *types.MessageItem) {
	if toUserID == 0 || item == nil {
		return
	}
	online, err := s.isUserOnline(ctx, toUserID)
	if err != nil {
		s.log.Warn(ctx, "检查在线状态失败，跳过推送", logger.Uint("user_id", toUserID), logger.Error(err))
		return
	}
	if online {
		return
	}
	devices, err := s.deviceRepo.ListEnabledByUserID(ctx, toUserID)
	if err != nil {
		s.log.Warn(ctx, "查询推送设备失败", logger.Uint("user_id", toUserID), logger.Error(err))
		return
	}
	if len(devices) == 0 {
		return
	}

	title := strings.TrimSpace(item.FromName)
	if title == "" {
		title = "新消息"
	}
	body := messagePreview(item.ContentType, item.Content)
	body = truncateRunes(body, 80)
	if body == "" {
		body = "[新消息]"
	}

	n := &push.Notification{
		Title: title,
		Body:  body,
		Data: map[string]string{
			"type":       "private",
			"messageId":  strconv.FormatUint(uint64(item.ID), 10),
			"fromUserId": strconv.FormatUint(uint64(item.FromUserID), 10),
			"peerId":     strconv.FormatUint(uint64(item.FromUserID), 10),
		},
	}

	for _, device := range devices {
		tokenVendor := device.Vendor
		token := device.PushToken
		go func(vendor, pushToken string) {
			bg := context.WithoutCancel(ctx)
			if sendErr := s.dispatch.Send(bg, vendor, pushToken, n); sendErr != nil {
				s.log.Warn(bg, "推送发送失败",
					logger.Uint("user_id", toUserID),
					logger.String("vendor", vendor),
					logger.Error(sendErr))
				var inv *push.InvalidTokenError
				if errors.As(sendErr, &inv) {
					_ = s.deviceRepo.DeleteByToken(bg, inv.Vendor, inv.Token)
				}
			}
		}(tokenVendor, token)
	}
}

func (s *DeviceService) isUserOnline(ctx context.Context, userID uint) (bool, error) {
	key := fmt.Sprintf(types.CacheWSPresenceKey, userID)
	all, err := s.cache.HGetAll(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return isUserPresenceOnline(all), nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}
