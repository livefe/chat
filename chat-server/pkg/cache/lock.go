package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end
`)

func (c *Cache) TryLock(ctx context.Context, key string, ttl time.Duration) (unlock func(context.Context) error, acquired bool, err error) {
	if c == nil || c.Client == nil {
		return nil, false, fmt.Errorf("缓存未初始化")
	}
	if key == "" {
		return nil, false, fmt.Errorf("锁 key 不能为空")
	}
	if ttl <= 0 {
		return nil, false, fmt.Errorf("锁 TTL 必须大于 0")
	}

	token := uuid.NewString()
	ok, err := c.Client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("获取锁失败: %w", err)
	}
	if !ok {
		return nil, false, nil
	}

	unlock = func(unlockCtx context.Context) error {
		if unlockCtx == nil {
			unlockCtx = context.Background()
		}
		if err := unlockScript.Run(unlockCtx, c.Client, []string{key}, token).Err(); err != nil {
			return fmt.Errorf("释放锁失败: %w", err)
		}
		return nil
	}
	return unlock, true, nil
}
