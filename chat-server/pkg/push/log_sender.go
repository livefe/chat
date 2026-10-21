package push

import (
	"context"
	"log/slog"
)

type LogSender struct {
	vendor string
}

func NewLogSender(vendor string) *LogSender {
	return &LogSender{vendor: vendor}
}

func (s *LogSender) Vendor() string { return s.vendor }

func (s *LogSender) Send(ctx context.Context, token string, n *Notification) error {
	slog.InfoContext(ctx, "push stub send",
		"vendor", s.vendor,
		"token_suffix", tokenSuffix(token),
		"title", n.Title,
		"body", n.Body,
		"data", n.Data,
	)
	return nil
}

func tokenSuffix(token string) string {
	if len(token) <= 8 {
		return token
	}
	return "..." + token[len(token)-8:]
}
