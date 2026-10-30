package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
)

const NullSentinel = "__NULL__"

type QueryOptions struct {
	CacheKey          string
	EmptyRecordExpire time.Duration
	DataExpire        time.Duration
}

type QueryFunc func(ctx context.Context) (any, error)

func Query(
	ctx context.Context,
	cache *Cache,
	options QueryOptions,
	queryFunc QueryFunc,
	result any,
) error {
	cachedData, err := cache.Client.Get(ctx, options.CacheKey).Result()
	if err == nil {
		if cachedData == NullSentinel {
			return gorm.ErrRecordNotFound
		}
		if err = json.Unmarshal([]byte(cachedData), result); err == nil {
			return nil
		}
	}

	queryResult, err := queryFunc(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = cache.Client.Set(ctx, options.CacheKey, NullSentinel, options.EmptyRecordExpire).Err()
		}
		return err
	}

	resultJSON, err := json.Marshal(queryResult)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	_ = cache.Client.Set(ctx, options.CacheKey, resultJSON, options.DataExpire).Err()

	if assignResult(result, queryResult) {
		return nil
	}
	if err := json.Unmarshal(resultJSON, result); err != nil {
		return fmt.Errorf("反序列化数据失败: %w", err)
	}
	return nil
}

func assignResult(dst, src any) bool {
	if dst == nil || src == nil {
		return false
	}
	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Ptr || dv.IsNil() {
		return false
	}
	de := dv.Elem()
	sv := reflect.ValueOf(src)
	if sv.Kind() == reflect.Ptr {
		if sv.IsNil() {
			return false
		}
		if sv.Elem().Type().AssignableTo(de.Type()) {
			de.Set(sv.Elem())
			return true
		}
		return false
	}
	if sv.Type().AssignableTo(de.Type()) {
		de.Set(sv)
		return true
	}
	return false
}
