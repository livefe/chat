package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Client interface {
	Upload(objectKey string, r io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, objectKey string) error
}

type Config struct {
	Provider       string        `mapstructure:"provider"`
	MaxSizeMB      int           `mapstructure:"max_size_mb"`
	MaxVideoSizeMB int           `mapstructure:"max_video_size_mb"`
	MaxAudioSizeMB int           `mapstructure:"max_audio_size_mb"`
	PathPrefix     string        `mapstructure:"path_prefix"`
	OrphanTTLHours int           `mapstructure:"orphan_ttl_hours"`
	Tencent        TencentConfig `mapstructure:"tencent"`
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Provider) == "" {
		return fmt.Errorf("storage.provider 未配置")
	}
	p := strings.ToLower(strings.TrimSpace(c.Provider))
	if p != "local" && p != "tencent" {
		return fmt.Errorf("storage.provider 不支持: %s", c.Provider)
	}
	if c.MaxSizeMB <= 0 {
		return fmt.Errorf("storage.max_size_mb 未配置")
	}
	if c.MaxVideoSizeMB <= 0 {
		return fmt.Errorf("storage.max_video_size_mb 未配置")
	}
	if c.MaxAudioSizeMB <= 0 {
		return fmt.Errorf("storage.max_audio_size_mb 未配置")
	}
	if strings.TrimSpace(c.PathPrefix) == "" {
		return fmt.Errorf("storage.path_prefix 未配置")
	}
	if c.OrphanTTLHours <= 0 {
		return fmt.Errorf("storage.orphan_ttl_hours 未配置")
	}
	if p == "tencent" {
		if strings.TrimSpace(c.Tencent.DefaultBucket) == "" {
			return fmt.Errorf("storage.tencent.default_bucket 未配置")
		}
		if strings.TrimSpace(c.Tencent.SecretID) == "" {
			return fmt.Errorf("storage.tencent.secret_id 未配置")
		}
		if strings.TrimSpace(c.Tencent.SecretKey) == "" {
			return fmt.Errorf("storage.tencent.secret_key 未配置")
		}
		if strings.TrimSpace(c.Tencent.Region) == "" {
			return fmt.Errorf("storage.tencent.region 未配置")
		}
	}
	return nil
}

func New(cfg *Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	p := provider(cfg)
	switch p {
	case "local":
		return newLocal()
	case "tencent":
		return newTencent(&cfg.Tencent)
	default:
		return nil, fmt.Errorf("不支持的对象存储: %s", cfg.Provider)
	}
}

func provider(cfg *Config) string {
	if cfg == nil {
		return ""
	}
	p := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch p {
	case "", "off", "none", "disabled":
		return ""
	default:
		return p
	}
}
