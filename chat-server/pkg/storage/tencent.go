package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	tcos "github.com/tencentyun/cos-go-sdk-v5"
)

type TencentConfig struct {
	SecretID  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`
	Region    string `mapstructure:"region"`
	Bucket    string `mapstructure:"bucket"`
	Domain    string `mapstructure:"domain"`
}

func (c *TencentConfig) Validate() error {
	if strings.TrimSpace(c.SecretID) == "" {
		return fmt.Errorf("storage.tencent.secret_id 未配置")
	}
	if strings.TrimSpace(c.SecretKey) == "" {
		return fmt.Errorf("storage.tencent.secret_key 未配置")
	}
	if strings.TrimSpace(c.Region) == "" {
		return fmt.Errorf("storage.tencent.region 未配置")
	}
	if strings.TrimSpace(c.Bucket) == "" {
		return fmt.Errorf("storage.tencent.bucket 未配置")
	}
	if strings.Contains(c.Domain, "://") || strings.Contains(c.Domain, "/") {
		return fmt.Errorf("storage.tencent.domain 只填域名，不含协议与路径")
	}
	return nil
}

type tencentClient struct {
	sdk     *tcos.Client
	baseURL string
}

func newTencent(cfg *TencentConfig) (Client, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("解析腾讯云 COS 地址失败: %w", err)
	}
	baseURL := bucketURL.String()
	if domain := strings.TrimSpace(cfg.Domain); domain != "" {
		baseURL = "https://" + domain
	}
	sdk := tcos.NewClient(&tcos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &tcos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})
	return &tencentClient{sdk: sdk, baseURL: baseURL}, nil
}

func (c *tencentClient) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	key, err := cleanKey(key)
	if err != nil {
		return err
	}
	header := &tcos.ObjectPutHeaderOptions{ContentType: contentType}
	if size > 0 {
		header.ContentLength = size
	}
	if _, err := c.sdk.Object.Put(ctx, key, r, &tcos.ObjectPutOptions{ObjectPutHeaderOptions: header}); err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	return nil
}

func (c *tencentClient) Delete(ctx context.Context, key string) error {
	key, err := cleanKey(key)
	if err != nil {
		return err
	}
	if _, err := c.sdk.Object.Delete(ctx, key); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

func (c *tencentClient) URL(key string) string {
	key, err := cleanKey(key)
	if err != nil {
		return ""
	}
	return c.baseURL + "/" + key
}
