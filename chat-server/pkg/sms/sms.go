package sms

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type SMS interface {
	SendSMS(ctx context.Context, phoneNumbers, signName, templateCode, templateParam string) (*SendResponse, error)
}

type SendResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	BizID     string `json:"biz_id"`
	RequestID string `json:"request_id"`
}

type AliyunConfig struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`
	RegionID        string `mapstructure:"region_id"`
}

type Config struct {
	Provider string       `mapstructure:"provider"`
	Aliyun   AliyunConfig `mapstructure:"aliyun"`
}

func (c *Config) Validate() error {
	provider := strings.TrimSpace(c.Provider)
	if provider == "" {
		return fmt.Errorf("sms.provider 未配置")
	}
	switch provider {
	case "mock":
		return nil
	case "aliyun":
		if strings.TrimSpace(c.Aliyun.AccessKeyID) == "" {
			return fmt.Errorf("sms.aliyun.access_key_id 未配置")
		}
		if strings.TrimSpace(c.Aliyun.AccessKeySecret) == "" {
			return fmt.Errorf("sms.aliyun.access_key_secret 未配置")
		}
		if strings.TrimSpace(c.Aliyun.Endpoint) == "" {
			return fmt.Errorf("sms.aliyun.endpoint 未配置")
		}
		if strings.TrimSpace(c.Aliyun.RegionID) == "" {
			return fmt.Errorf("sms.aliyun.region_id 未配置")
		}
		return nil
	default:
		return fmt.Errorf("sms.provider 不支持: %s", provider)
	}
}

type MockSMS struct{}

func (m *MockSMS) SendSMS(ctx context.Context, phoneNumbers, signName, templateCode, templateParam string) (*SendResponse, error) {
	return &SendResponse{
		Code:      "OK",
		Message:   "发送成功（mock，验证码固定 111111）",
		BizID:     fmt.Sprintf("mock_%d", time.Now().UnixNano()),
		RequestID: fmt.Sprintf("mock_%d", time.Now().UnixNano()),
	}, nil
}

func NewSMS(config Config) (SMS, error) {
	switch config.Provider {
	case "aliyun":
		return NewAliyun(config.Aliyun)
	case "mock":
		return &MockSMS{}, nil
	default:
		return nil, fmt.Errorf("不支持的短信服务商: %s", config.Provider)
	}
}
