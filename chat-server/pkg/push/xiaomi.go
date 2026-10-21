package push

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const xiaomiPushURL = "https://api.xmpush.xiaomi.com/v3/message/regid"

type XiaomiConfig struct {
	AppSecret   string `mapstructure:"app_secret"`
	PackageName string `mapstructure:"package_name"`
	ChannelID   string `mapstructure:"channel_id"`
}

func (c *XiaomiConfig) Validate() error {
	secret := strings.TrimSpace(c.AppSecret)
	pkg := strings.TrimSpace(c.PackageName)
	if secret == "" && pkg == "" {
		return nil
	}
	if secret == "" {
		return fmt.Errorf("push.xiaomi.app_secret 未配置")
	}
	if pkg == "" {
		return fmt.Errorf("push.xiaomi.package_name 未配置")
	}
	return nil
}

type XiaomiSender struct {
	cfg    XiaomiConfig
	client *http.Client
}

func NewXiaomiSender(cfg XiaomiConfig) *XiaomiSender {
	return &XiaomiSender{
		cfg: cfg,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (s *XiaomiSender) Vendor() string { return "xiaomi" }

func (s *XiaomiSender) Send(ctx context.Context, token string, n *Notification) error {
	if strings.TrimSpace(s.cfg.AppSecret) == "" {
		return fmt.Errorf("xiaomi push: app_secret 未配置")
	}
	if strings.TrimSpace(s.cfg.PackageName) == "" {
		return fmt.Errorf("xiaomi push: package_name 未配置")
	}
	if strings.TrimSpace(token) == "" || n == nil {
		return fmt.Errorf("xiaomi push: token/通知为空")
	}

	payloadBytes, _ := json.Marshal(n.Data)
	form := url.Values{}
	form.Set("registration_id", token)
	form.Set("restricted_package_name", s.cfg.PackageName)
	form.Set("title", truncate(n.Title, 50))
	form.Set("description", truncate(n.Body, 128))
	form.Set("payload", string(payloadBytes))
	form.Set("pass_through", "0")
	form.Set("notify_type", "-1")
	form.Set("extra.notify_effect", "1")
	form.Set("extra.notify_foreground", "1")
	if s.cfg.ChannelID != "" {
		form.Set("extra.channel_id", s.cfg.ChannelID)
	}

	if peerID := n.Data["peerId"]; peerID != "" {
		intent := fmt.Sprintf(
			"intent:#Intent;component=%s/.MainActivity;l.peerId=%s;end",
			s.cfg.PackageName,
			peerID,
		)
		form.Set("extra.notify_effect", "2")
		form.Set("extra.intent_uri", intent)
		form.Set("extra.peerId", peerID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, xiaomiPushURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "key="+s.cfg.AppSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("xiaomi push http: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	var result struct {
		Code        int64           `json:"code"`
		Result      string          `json:"result"`
		Description string          `json:"description"`
		Data        json.RawMessage `json:"data"`
		Reason      string          `json:"reason"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("xiaomi push 响应解析失败 status=%d body=%s", resp.StatusCode, string(body))
	}

	if result.Code != 0 {

		if result.Code == 20301 || result.Code == 20302 || strings.Contains(strings.ToLower(result.Description+result.Reason), "invalid") {
			return &InvalidTokenError{Vendor: "xiaomi", Token: token, Err: fmt.Errorf("code=%d %s", result.Code, result.Description)}
		}
		return fmt.Errorf("xiaomi push failed code=%d result=%s desc=%s", result.Code, result.Result, result.Description)
	}
	return nil
}

func truncate(s string, max int) string {
	r := []rune(strings.TrimSpace(s))
	if max <= 0 || len(r) <= max {
		return string(r)
	}
	return string(r[:max])
}

func BuildDispatchers(xiaomi XiaomiConfig) *Dispatcher {
	senders := []Sender{
		NewLogSender("huawei"),
		NewLogSender("honor"),
		NewLogSender("oppo"),
		NewLogSender("vivo"),
		NewLogSender("fcm"),
		NewLogSender("other"),
	}
	if strings.TrimSpace(xiaomi.AppSecret) != "" && strings.TrimSpace(xiaomi.PackageName) != "" {
		senders = append(senders, NewXiaomiSender(xiaomi))
	} else {
		senders = append(senders, NewLogSender("xiaomi"))
	}
	return NewDispatcher(senders...)
}
