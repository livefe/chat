package service

import (
	"context"
	"encoding/json"
	"fmt"

	"app/internal/model"
	"app/internal/repository"
	"app/internal/types"
	"app/pkg/cache"
)

type UserStore struct {
	repository.UserRepository
	cache *cache.Cache
}

func NewUserStore(repo repository.UserRepository, c *cache.Cache) *UserStore {
	return &UserStore{UserRepository: repo, cache: c}
}

func (s *UserStore) Create(ctx context.Context, user *model.User) error {
	if err := s.UserRepository.Create(ctx, user); err != nil {
		return err
	}
	s.invalidate(ctx, user.ID, user.Mobile)
	return nil
}

func (s *UserStore) Update(ctx context.Context, user *model.User) error {
	if err := s.UserRepository.Update(ctx, user); err != nil {
		return err
	}
	s.invalidate(ctx, user.ID, user.Mobile)
	return nil
}

func (s *UserStore) UpdateMobile(ctx context.Context, user *model.User, oldMobile string) error {
	if err := s.UserRepository.Update(ctx, user); err != nil {
		return err
	}
	s.invalidate(ctx, user.ID, oldMobile, user.Mobile)
	return nil
}

func (s *UserStore) Deactivate(ctx context.Context, user *model.User, oldMobile string) error {
	if err := s.UserRepository.Deactivate(ctx, user); err != nil {
		return err
	}
	s.invalidate(ctx, user.ID, oldMobile, user.Mobile)
	return nil
}

func (s *UserStore) GetByIDWithCache(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := cache.Query(ctx, s.cache, cache.QueryOptions{
		CacheKey:          fmt.Sprintf(types.CacheUserByIDKey, id),
		EmptyRecordExpire: types.Expire5m,
		DataExpire:        types.Expire1h,
	}, func(ctx context.Context) (any, error) {
		return s.UserRepository.GetByID(ctx, id)
	}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) GetByIDsWithCache(ctx context.Context, ids []uint) ([]*model.User, error) {
	uniq := uniqueNonZeroIDs(ids)
	if len(uniq) == 0 {
		return nil, nil
	}

	keys := make([]string, len(uniq))
	for i, id := range uniq {
		keys[i] = fmt.Sprintf(types.CacheUserByIDKey, id)
	}
	vals, err := s.cache.MGet(ctx, keys...).Result()
	if err != nil {
		return s.UserRepository.GetByIDs(ctx, uniq)
	}

	out := make([]*model.User, 0, len(uniq))
	miss := make([]uint, 0)
	for i, v := range vals {
		if v == nil {
			miss = append(miss, uniq[i])
			continue
		}
		raw, ok := v.(string)
		if !ok || raw == "" || raw == "__NULL__" {
			continue
		}
		var user model.User
		if err := json.Unmarshal([]byte(raw), &user); err != nil {
			miss = append(miss, uniq[i])
			continue
		}
		out = append(out, &user)
	}
	if len(miss) == 0 {
		return out, nil
	}

	loaded, err := s.UserRepository.GetByIDs(ctx, miss)
	if err != nil {
		return out, err
	}
	if len(loaded) == 0 {
		return out, nil
	}
	pipe := s.cache.Pipeline()
	for _, u := range loaded {
		b, mErr := json.Marshal(u)
		if mErr != nil {
			continue
		}
		pipe.Set(ctx, fmt.Sprintf(types.CacheUserByIDKey, u.ID), b, types.Expire1h)
		out = append(out, u)
	}
	_, _ = pipe.Exec(ctx)
	return out, nil
}

func (s *UserStore) invalidate(ctx context.Context, id uint, mobiles ...string) {
	keys := make([]string, 0, len(mobiles)+1)
	keys = append(keys, fmt.Sprintf(types.CacheUserByIDKey, id))
	for _, m := range mobiles {
		keys = append(keys, fmt.Sprintf(types.CacheUserByMobileKey, m))
	}
	_ = s.cache.Del(ctx, keys...).Err()
}

func uniqueNonZeroIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	uniq := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}
	return uniq
}
