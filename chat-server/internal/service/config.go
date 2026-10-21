package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

const cacheConfigKey = "app:config:%s"

const configKeyAd = "ad"

const configRemarkAd = "发现页广告投放规则"

type ConfigService struct {
	log   *logger.Logger
	cache *cache.Cache
	repo  repository.ConfigRepository

	mu sync.RWMutex
	ad types.AdminAdSetting
}

func NewConfigService(log *logger.Logger, c *cache.Cache, repo repository.ConfigRepository) *ConfigService {
	s := &ConfigService{
		log:   log,
		cache: c,
		repo:  repo,
	}
	s.reload(context.Background())
	return s
}

func (s *ConfigService) GetAd() types.AdminAdSetting {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ad
}

func (s *ConfigService) SaveAd(ctx context.Context, in types.AdminAdSetting) error {
	normalized, err := normalizeAdSetting(in)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}
	if err := s.repo.Upsert(ctx, configKeyAd, string(raw), configRemarkAd); err != nil {
		s.log.Error(ctx, "保存投放设置失败", logger.Error(err))
		return fmt.Errorf("保存失败: %w", err)
	}
	_ = s.cache.Del(ctx, fmt.Sprintf(cacheConfigKey, configKeyAd)).Err()
	s.mu.Lock()
	s.ad = normalized
	s.mu.Unlock()
	return nil
}

func (s *ConfigService) getByKey(ctx context.Context, key string) (*model.Config, error) {
	var row model.Config
	err := cache.Query(ctx, s.cache, cache.QueryOptions{
		CacheKey:          fmt.Sprintf(cacheConfigKey, key),
		EmptyRecordExpire: types.Expire5m,
		DataExpire:        types.Expire1d,
	}, func(ctx context.Context) (any, error) {
		return s.repo.GetByKey(ctx, key)
	}, &row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *ConfigService) reload(ctx context.Context) {
	row, err := s.getByKey(ctx, configKeyAd)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			s.log.Warn(ctx, "读取投放设置失败", logger.Error(err))
		}
		return
	}
	var parsed types.AdminAdSetting
	if err := json.Unmarshal([]byte(row.Value), &parsed); err != nil {
		s.log.Warn(ctx, "投放设置 JSON 无效", logger.Error(err))
		return
	}
	normalized, err := normalizeAdSetting(parsed)
	if err != nil {
		s.log.Warn(ctx, "投放设置不合法", logger.Error(err))
		return
	}
	s.mu.Lock()
	s.ad = normalized
	s.mu.Unlock()
}

func normalizeAdSetting(in types.AdminAdSetting) (types.AdminAdSetting, error) {
	if in.FirstSlotAfter < 1 || in.FirstSlotAfter > 100 {
		return types.AdminAdSetting{}, errs.Invalid("首条广告位置需在 1 到 100 之间")
	}
	if in.SlotInterval < 1 || in.SlotInterval > 100 {
		return types.AdminAdSetting{}, errs.Invalid("广告间隔需在 1 到 100 之间")
	}
	if in.MaxActivePerAdvertiser < 1 || in.MaxActivePerAdvertiser > 50 {
		return types.AdminAdSetting{}, errs.Invalid("同一商家投放上限需在 1 到 50 之间")
	}
	if in.LowCTRThreshold <= 0 || in.LowCTRThreshold > 1 {
		return types.AdminAdSetting{}, errs.Invalid("低点击率阈值需大于 0 且不超过 1")
	}
	return in, nil
}
