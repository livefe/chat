package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

type brokerEnvelope struct {
	Origin  string `json:"o"`
	Payload []byte `json:"p"`
}

type RedisBroker struct {
	client     *redis.Client
	instanceID string

	mu     sync.Mutex
	subs   []*redis.PubSub
	ctx    context.Context
	cancel context.CancelFunc
	closed bool
}

func NewRedisBroker(client *redis.Client, instanceID string) *RedisBroker {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisBroker{
		client:     client,
		instanceID: instanceID,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (b *RedisBroker) InstanceID() string { return b.instanceID }

func (b *RedisBroker) Publish(ctx context.Context, channel string, payload []byte) error {
	data, err := json.Marshal(brokerEnvelope{Origin: b.instanceID, Payload: payload})
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	return b.client.Publish(ctx, channel, data).Err()
}

func (b *RedisBroker) Subscribe(pattern string, handler func(channel string, payload []byte)) error {
	ps := b.client.PSubscribe(b.ctx, pattern)
	if _, err := ps.Receive(b.ctx); err != nil {
		_ = ps.Close()
		return fmt.Errorf("psubscribe %q: %w", pattern, err)
	}

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		_ = ps.Close()
		return errors.New("websocket: redis broker already closed")
	}
	b.subs = append(b.subs, ps)
	b.mu.Unlock()

	go func() {
		for msg := range ps.Channel() {
			var env brokerEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				continue
			}
			if env.Origin == b.instanceID {
				continue
			}
			handler(msg.Channel, env.Payload)
		}
	}()
	return nil
}

func (b *RedisBroker) Close(_ context.Context) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	subs := b.subs
	b.subs = nil
	b.mu.Unlock()

	b.cancel()
	var firstErr error
	for _, ps := range subs {
		if err := ps.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
