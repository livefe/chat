package websocket

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gorillaws "github.com/gorilla/websocket"
)

type ConnConfig struct {
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	ReadDeadline      time.Duration `mapstructure:"read_deadline"`
	WriteDeadline     time.Duration `mapstructure:"write_deadline"`
	SendBuffer        int           `mapstructure:"send_buffer"`
	MaxMessageSize    int64         `mapstructure:"max_message_size"`
}

type Config struct {
	ConnConfig      `mapstructure:",squash"`
	MaxConnections  int      `mapstructure:"max_connections"`
	ReadBufferSize  int      `mapstructure:"read_buffer_size"`
	WriteBufferSize int      `mapstructure:"write_buffer_size"`
	AllowOrigins    []string `mapstructure:"allow_origins"`
}

func (c *Config) Validate() error {
	if c.MaxConnections <= 0 {
		return fmt.Errorf("websocket.max_connections 未配置")
	}
	if c.HeartbeatInterval <= 0 {
		return fmt.Errorf("websocket.heartbeat_interval 未配置")
	}
	if c.ReadDeadline <= 0 {
		return fmt.Errorf("websocket.read_deadline 未配置")
	}
	if c.WriteDeadline <= 0 {
		return fmt.Errorf("websocket.write_deadline 未配置")
	}
	if c.SendBuffer <= 0 {
		return fmt.Errorf("websocket.send_buffer 未配置")
	}
	if c.MaxMessageSize <= 0 {
		return fmt.Errorf("websocket.max_message_size 未配置")
	}
	if c.ReadBufferSize <= 0 {
		return fmt.Errorf("websocket.read_buffer_size 未配置")
	}
	if c.WriteBufferSize <= 0 {
		return fmt.Errorf("websocket.write_buffer_size 未配置")
	}
	if len(c.AllowOrigins) == 0 {
		return fmt.Errorf("websocket.allow_origins 未配置")
	}
	for _, origin := range c.AllowOrigins {
		if strings.TrimSpace(origin) == "" {
			return fmt.Errorf("websocket.allow_origins 含空项")
		}
	}
	return nil
}

func NewUpgrader(cfg *Config, onOriginRejected func(origin string)) *gorillaws.Upgrader {
	allowed := make(map[string]struct{}, len(cfg.AllowOrigins))
	allowAll := false
	for _, o := range cfg.AllowOrigins {
		if o == "" {
			continue
		}
		if o == "*" {
			allowAll = true
			continue
		}
		allowed[o] = struct{}{}
	}
	return &gorillaws.Upgrader{
		ReadBufferSize:  cfg.ReadBufferSize,
		WriteBufferSize: cfg.WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			if origin == "" {
				return true
			}

			if allowAll {
				return true
			}
			if _, ok := allowed[origin]; ok {
				return true
			}
			if onOriginRejected != nil {
				onOriginRejected(origin)
			}
			return false
		},
	}
}

type Handler interface {
	OnOpen(c *Conn)
	OnMessage(c *Conn, payload []byte)
	OnClose(c *Conn)
}

var inboundBusyPayload = []byte(`{"type":"error","message":"处理繁忙，请稍后重试"}`)

type Conn struct {
	raw     *gorillaws.Conn
	cfg     ConnConfig
	handler Handler

	send     chan []byte
	inbound  chan []byte
	done     chan struct{}
	doneOnce sync.Once
	closed   atomic.Bool

	attrs sync.Map
}

func NewConn(raw *gorillaws.Conn, cfg ConnConfig, h Handler) *Conn {
	inboundSize := cfg.SendBuffer
	if inboundSize < 32 {
		inboundSize = 32
	}
	return &Conn{
		raw:     raw,
		cfg:     cfg,
		handler: h,
		send:    make(chan []byte, cfg.SendBuffer),
		inbound: make(chan []byte, inboundSize),
		done:    make(chan struct{}),
	}
}

func (c *Conn) Start() {
	rdl := time.Now().Add(c.cfg.ReadDeadline)
	c.raw.SetReadLimit(c.cfg.MaxMessageSize)
	_ = c.raw.SetReadDeadline(rdl)
	c.raw.SetPongHandler(func(string) error {
		return c.raw.SetReadDeadline(time.Now().Add(c.cfg.ReadDeadline))
	})
	if c.handler != nil {
		c.handler.OnOpen(c)
	}
	go c.writePump()
	go c.processPump()
	go c.readPump()
}

func (c *Conn) readPump() {
	defer c.cleanup()
	for {
		_, payload, err := c.raw.ReadMessage()
		if err != nil {
			return
		}

		_ = c.raw.SetReadDeadline(time.Now().Add(c.cfg.ReadDeadline))
		if c.handler == nil {
			continue
		}

		buf := make([]byte, len(payload))
		copy(buf, payload)

		select {
		case c.inbound <- buf:
		case <-c.done:
			return
		default:
			_ = c.Send(inboundBusyPayload)
		}
	}
}

func (c *Conn) processPump() {
	for {
		select {
		case payload, ok := <-c.inbound:
			if !ok {
				return
			}
			if c.handler == nil || c.closed.Load() {
				continue
			}
			func() {
				defer func() { _ = recover() }()
				c.handler.OnMessage(c, payload)
			}()
		case <-c.done:
			return
		}
	}
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(c.cfg.HeartbeatInterval)
	defer func() {
		ticker.Stop()
		c.cleanup()
	}()
	wdl := c.cfg.WriteDeadline
	for {
		select {
		case payload, ok := <-c.send:
			if !ok {
				_ = c.raw.WriteMessage(gorillaws.CloseMessage, nil)
				return
			}
			_ = c.raw.SetWriteDeadline(time.Now().Add(wdl))
			if err := c.raw.WriteMessage(gorillaws.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.raw.SetWriteDeadline(time.Now().Add(wdl))
			if err := c.raw.WriteMessage(gorillaws.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (c *Conn) Send(payload []byte) error {
	if c.closed.Load() {
		return errors.New("websocket: conn closed")
	}
	select {
	case c.send <- payload:
		return nil
	default:

		c.Close()
		return errors.New("websocket: send buffer full")
	}
}

func (c *Conn) Set(key string, value any)  { c.attrs.Store(key, value) }
func (c *Conn) Get(key string) (any, bool) { return c.attrs.Load(key) }

func (c *Conn) RemoteAddr() string { return c.raw.RemoteAddr().String() }

func (c *Conn) Close() {
	c.doneOnce.Do(func() { close(c.done) })
	_ = c.raw.Close()
}

func (c *Conn) cleanup() {
	if !c.closed.CompareAndSwap(false, true) {
		return
	}
	c.doneOnce.Do(func() { close(c.done) })
	_ = c.raw.Close()
	if c.handler != nil {
		c.handler.OnClose(c)
	}
}
