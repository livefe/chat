package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type Client interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, key string) error
	URL(key string) string
}

const (
	ProviderLocal   = "local"
	ProviderTencent = "tencent"
)

type Config struct {
	Provider string        `mapstructure:"provider"`
	Tencent  TencentConfig `mapstructure:"tencent"`
}

func (c *Config) Validate() error {
	switch c.provider() {
	case ProviderLocal:
		return nil
	case ProviderTencent:
		return c.Tencent.Validate()
	case "":
		return fmt.Errorf("storage.provider 未配置")
	default:
		return fmt.Errorf("storage.provider 不支持: %s（允许 local / tencent）", c.Provider)
	}
}

func (c *Config) provider() string {
	return strings.ToLower(strings.TrimSpace(c.Provider))
}

func New(cfg *Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if cfg.provider() == ProviderTencent {
		return newTencent(&cfg.Tencent)
	}
	return newLocal()
}

func cleanKey(key string) (string, error) {
	key = strings.Trim(strings.ReplaceAll(strings.TrimSpace(key), "\\", "/"), "/")
	if key == "" {
		return "", fmt.Errorf("对象键为空")
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return "", fmt.Errorf("非法对象键: %s", key)
		}
	}
	return key, nil
}
