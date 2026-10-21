package websocket

import (
	"sync"
)

type Hub[ID comparable] struct {
	mu         sync.RWMutex
	byID       map[ID]map[*Conn]struct{}
	byConn     map[*Conn]ID
	topics     map[string]map[*Conn]struct{}
	connTopics map[*Conn]map[string]struct{}
}

func NewHub[ID comparable]() *Hub[ID] {
	return &Hub[ID]{
		byID:       make(map[ID]map[*Conn]struct{}),
		byConn:     make(map[*Conn]ID),
		topics:     make(map[string]map[*Conn]struct{}),
		connTopics: make(map[*Conn]map[string]struct{}),
	}
}

func (h *Hub[ID]) Register(id ID, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.registerLocked(id, c)
}

func (h *Hub[ID]) RegisterIfUnder(id ID, c *Conn, max int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if max > 0 && len(h.byConn) >= max {
		return false
	}
	h.registerLocked(id, c)
	return true
}

func (h *Hub[ID]) registerLocked(id ID, c *Conn) {
	set, ok := h.byID[id]
	if !ok {
		set = make(map[*Conn]struct{}, 2)
		h.byID[id] = set
	}
	set[c] = struct{}{}
	h.byConn[c] = id
}

func (h *Hub[ID]) ConnCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byConn)
}

func (h *Hub[ID]) Unregister(c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if id, ok := h.byConn[c]; ok {
		if set := h.byID[id]; set != nil {
			delete(set, c)
			if len(set) == 0 {
				delete(h.byID, id)
			}
		}
		delete(h.byConn, c)
	}

	for topic := range h.connTopics[c] {
		if set := h.topics[topic]; set != nil {
			delete(set, c)
			if len(set) == 0 {
				delete(h.topics, topic)
			}
		}
	}
	delete(h.connTopics, c)
}

func (h *Hub[ID]) Subscribe(c *Conn, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.topics[topic]
	if !ok {
		set = make(map[*Conn]struct{}, 4)
		h.topics[topic] = set
	}
	set[c] = struct{}{}
	ct, ok := h.connTopics[c]
	if !ok {
		ct = make(map[string]struct{}, 2)
		h.connTopics[c] = ct
	}
	ct[topic] = struct{}{}
}

func (h *Hub[ID]) Unsubscribe(c *Conn, topic string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set, ok := h.topics[topic]
	if !ok {
		return
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.topics, topic)
	}
	if ct := h.connTopics[c]; ct != nil {
		delete(ct, topic)
		if len(ct) == 0 {
			delete(h.connTopics, c)
		}
	}
}

func (h *Hub[ID]) CountByID(id ID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byID[id])
}

func (h *Hub[ID]) HasID(id ID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byID[id]) > 0
}

func (h *Hub[ID]) ConnsByID(id ID) []*Conn {
	h.mu.RLock()
	set := h.byID[id]
	targets := make([]*Conn, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()
	return targets
}

func (h *Hub[ID]) ConnsByTopic(topic string) []*Conn {
	h.mu.RLock()
	set := h.topics[topic]
	targets := make([]*Conn, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()
	return targets
}

func (h *Hub[ID]) SendToID(id ID, payload []byte) int {
	h.mu.RLock()
	set := h.byID[id]
	targets := make([]*Conn, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	n := 0
	for _, c := range targets {
		if err := c.Send(payload); err == nil {
			n++
		}
	}
	return n
}

func (h *Hub[ID]) SendToIDs(ids []ID, payload []byte) int {
	if len(ids) == 0 || len(payload) == 0 {
		return 0
	}
	h.mu.RLock()
	targets := make([]*Conn, 0)
	for _, id := range ids {
		for c := range h.byID[id] {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()

	n := 0
	for _, c := range targets {
		if err := c.Send(payload); err == nil {
			n++
		}
	}
	return n
}

func (h *Hub[ID]) SendToTopic(topic string, payload []byte, excludeConn *Conn) int {
	h.mu.RLock()
	set := h.topics[topic]
	targets := make([]*Conn, 0, len(set))
	for c := range set {
		if c == excludeConn {
			continue
		}
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	n := 0
	for _, c := range targets {
		if err := c.Send(payload); err == nil {
			n++
		}
	}
	return n
}

func (h *Hub[ID]) CloseByID(id ID) int {
	h.mu.RLock()
	set := h.byID[id]
	targets := make([]*Conn, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	n := 0
	for _, c := range targets {
		c.Close()
		n++
	}
	return n
}

func (h *Hub[ID]) CloseAll() int {
	h.mu.RLock()
	targets := make([]*Conn, 0, len(h.byConn))
	for c := range h.byConn {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	n := 0
	for _, c := range targets {
		c.Close()
		n++
	}
	return n
}

func (h *Hub[ID]) SnapshotByUser() map[ID][]*Conn {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[ID][]*Conn, len(h.byID))
	for id, set := range h.byID {
		conns := make([]*Conn, 0, len(set))
		for c := range set {
			conns = append(conns, c)
		}
		out[id] = conns
	}
	return out
}
