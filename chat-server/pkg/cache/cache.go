package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxRetries      int           `mapstructure:"max_retries"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	PoolTimeout     time.Duration `mapstructure:"pool_timeout"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("cache.host 未配置")
	}
	if c.Port <= 0 {
		return fmt.Errorf("cache.port 未配置")
	}
	if c.PoolSize <= 0 {
		return fmt.Errorf("cache.pool_size 未配置")
	}
	if c.MinIdleConns <= 0 {
		return fmt.Errorf("cache.min_idle_conns 未配置")
	}
	if c.MaxRetries <= 0 {
		return fmt.Errorf("cache.max_retries 未配置")
	}
	if c.DialTimeout <= 0 {
		return fmt.Errorf("cache.dial_timeout 未配置")
	}
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("cache.read_timeout 未配置")
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("cache.write_timeout 未配置")
	}
	if c.PoolTimeout <= 0 {
		return fmt.Errorf("cache.pool_timeout 未配置")
	}
	if c.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("cache.conn_max_idle_time 未配置")
	}
	return nil
}

type Cache struct {
	*redis.Client
}

func Init(cfg *Config) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxRetries:      cfg.MaxRetries,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("缓存连接失败: %w", err)
	}

	return &Cache{Client: client}, nil
}

func (c *Cache) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}

	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("关闭缓存连接失败: %w", err)
	}

	return nil
}
