package push

import (
	"context"
	"fmt"
)

type Notification struct {
	Title string
	Body  string
	Data  map[string]string
}

type Sender interface {
	Vendor() string
	Send(ctx context.Context, token string, n *Notification) error
}

type Dispatcher struct {
	senders map[string]Sender
}

type InvalidTokenError struct {
	Vendor string
	Token  string
	Err    error
}

func (e *InvalidTokenError) Error() string {
	return fmt.Sprintf("push token invalid vendor=%s: %v", e.Vendor, e.Err)
}

func (e *InvalidTokenError) Unwrap() error { return e.Err }

func NewDispatcher(senders ...Sender) *Dispatcher {
	m := make(map[string]Sender, len(senders))
	for _, s := range senders {
		if s == nil {
			continue
		}
		m[s.Vendor()] = s
	}
	return &Dispatcher{senders: m}
}

func (d *Dispatcher) Send(ctx context.Context, vendor, token string, n *Notification) error {
	if token == "" || n == nil {
		return fmt.Errorf("push token 或通知为空")
	}
	s, ok := d.senders[vendor]
	if !ok {
		return fmt.Errorf("未配置厂商推送通道: %s", vendor)
	}
	return s.Send(ctx, token, n)
}
