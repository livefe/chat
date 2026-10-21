package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrExceeded = fmt.Errorf("rate limit exceeded")

func AllowSliding(ctx context.Context, rdb redis.Cmdable, key string, limit int, window time.Duration) error {
	if rdb == nil || limit <= 0 || window <= 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	minScore := strconv.FormatInt(now-window.Milliseconds(), 10)
	member := fmt.Sprintf("%d-%d", now, time.Now().UnixNano())

	pipe := rdb.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", minScore)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
	card := pipe.ZCard(ctx, key)
	pipe.Expire(ctx, key, window+time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil
	}
	if card.Val() > int64(limit) {
		return ErrExceeded
	}
	return nil
}
