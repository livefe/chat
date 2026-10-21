package sms

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyun struct {
	client *dysmsapi.Client
}

func NewAliyun(config AliyunConfig) (*Aliyun, error) {
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(config.AccessKeyID),
		AccessKeySecret: tea.String(config.AccessKeySecret),
		Endpoint:        tea.String(config.Endpoint),
		RegionId:        tea.String(config.RegionID),
	}

	client, err := dysmsapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建短信客户端失败: %w", err)
	}

	return &Aliyun{client: client}, nil
}

func (a *Aliyun) SendSMS(ctx context.Context, phoneNumbers, signName, templateCode, templateParam string) (*SendResponse, error) {
	request := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNumbers),
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(templateParam),
	}

	response, err := a.client.SendSms(request)
	if err != nil {
		return nil, fmt.Errorf("发送短信失败: %w", err)
	}

	result := &SendResponse{
		BizID:     *response.Body.BizId,
		RequestID: *response.Body.RequestId,
		Code:      *response.Body.Code,
		Message:   *response.Body.Message,
	}

	if *response.Body.Code != "OK" {
		return result, fmt.Errorf("发送短信失败: %s, %s", *response.Body.Code, *response.Body.Message)
	}

	return result, nil
}
