package service

import (
	"context"
	"encoding/json"
	"fmt"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/logger"
)

type employeePermCache struct {
	IsSuper bool                `json:"isSuper"`
	Perms   []*model.Permission `json:"perms"`
}

func (s *PermissionService) loadEmployeePermCache(ctx context.Context, employeeID uint) (*employeePermCache, bool) {
	if employeeID == 0 {
		return nil, false
	}
	raw, err := s.cache.Get(ctx, fmt.Sprintf(types.CacheEmployeePermKey, employeeID)).Bytes()
	if err != nil {
		return nil, false
	}
	var cached employeePermCache
	if err := json.Unmarshal(raw, &cached); err != nil {
		return nil, false
	}
	return &cached, true
}

func (s *PermissionService) storeEmployeePermCache(ctx context.Context, employeeID uint, isSuper bool, perms []*model.Permission) {
	if employeeID == 0 {
		return
	}
	raw, err := json.Marshal(employeePermCache{IsSuper: isSuper, Perms: perms})
	if err != nil {
		return
	}
	if err := s.cache.Set(ctx, fmt.Sprintf(types.CacheEmployeePermKey, employeeID), raw, types.Expire5m).Err(); err != nil {
		s.log.Warn(ctx, "缓存员工权限快照失败", logger.Uint("employee_id", employeeID), logger.Error(err))
	}
}

func (s *PermissionService) InvalidateEmployeePermCache(ctx context.Context, employeeID uint) {
	if employeeID == 0 {
		return
	}
	if err := s.cache.Del(ctx, fmt.Sprintf(types.CacheEmployeePermKey, employeeID)).Err(); err != nil {
		s.log.Warn(ctx, "清理员工权限缓存失败", logger.Uint("employee_id", employeeID), logger.Error(err))
	}
}

func (s *PermissionService) InvalidateAllEmployeePermCache(ctx context.Context) {
	s.invalidateAllEmployeePermCache(ctx)
}

func (s *PermissionService) invalidateAllEmployeePermCache(ctx context.Context) {
	var keys []string
	iter := s.cache.Scan(ctx, 0, types.CacheEmployeePermPattern, 200).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		s.log.Warn(ctx, "扫描员工权限缓存失败", logger.Error(err))
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := s.cache.Del(ctx, keys...).Err(); err != nil {
		s.log.Warn(ctx, "清理员工权限缓存失败", logger.Error(err))
	}
}
