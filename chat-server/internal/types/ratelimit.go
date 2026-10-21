package types

import "time"

const (
	UserSendRateLimit  = 8
	UserSendRateWindow = time.Second

	IMSendRateLimit  = 12
	IMSendRateWindow = time.Second

	CacheUserSendRateKey = cachePrefix + "rl:user:send:%d"
	CacheIMSendRateKey   = cachePrefix + "rl:im:send:%d"
)
