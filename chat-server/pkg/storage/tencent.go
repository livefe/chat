package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	tcos "github.com/tencentyun/cos-go-sdk-v5"
)

type TencentConfig struct {
	SecretID      string            `mapstructure:"secret_id"`
	SecretKey     string            `mapstructure:"secret_key"`
	Region        string            `mapstructure:"region"`
	DefaultBucket string            `mapstructure:"default_bucket"`
	UseDomainMap  bool              `mapstructure:"use_domain_map"`
	Buckets       map[string]string `mapstructure:"buckets"`
}

type tencentClient struct {
	cfg TencentConfig
}

func newTencent(cfg *TencentConfig) (Client, error) {
	if _, err := tencentSDK(*cfg, cfg.DefaultBucket); err != nil {
		return nil, err
	}
	return &tencentClient{cfg: *cfg}, nil
}

func (c *tencentClient) Upload(objectKey string, reader io.Reader, contentType string) (string, error) {
	client, err := tencentSDK(c.cfg, c.cfg.DefaultBucket)
	if err != nil {
		return "", err
	}
	opt := &tcos.ObjectPutOptions{}
	if contentType != "" {
		opt.ObjectPutHeaderOptions = &tcos.ObjectPutHeaderOptions{ContentType: contentType}
	}
	if _, err = client.Object.Put(context.Background(), objectKey, reader, opt); err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	return c.objectURL(c.cfg.DefaultBucket, objectKey), nil
}

func (c *tencentClient) Delete(ctx context.Context, objectKey string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	client, err := tencentSDK(c.cfg, c.cfg.DefaultBucket)
	if err != nil {
		return err
	}
	if _, err = client.Object.Delete(ctx, objectKey); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

func (c *tencentClient) objectURL(bucket, objectKey string) string {
	if c.cfg.UseDomainMap && c.cfg.Buckets != nil {
		if domain, ok := c.cfg.Buckets[bucket]; ok && domain != "" {
			return fmt.Sprintf("https://%s/%s", domain, objectKey)
		}
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", bucket, c.cfg.Region, objectKey)
}

func tencentSDK(cfg TencentConfig, bucket string) (*tcos.Client, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("解析腾讯云 COS URL 失败: %w", err)
	}
	return tcos.NewClient(&tcos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &tcos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	}), nil
}
