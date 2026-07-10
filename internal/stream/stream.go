// Package stream fans out real-time updates to connected web clients over SSE.
package stream

import "sync"

// Hub fans out published messages to every current subscriber. A slow
// subscriber gets messages dropped rather than blocking the publisher.
type Hub struct {
	mu   sync.Mutex
	subs map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[chan []byte]struct{})}
}

func (h *Hub) Subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
	return ch, unsubscribe
}

func (h *Hub) Publish(data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- data:
		default:
		}
	}
}

// ResourceMessage announces a raw bridge on/off change for id (a light, zone,
// or room), regardless of whether it's watched — used to keep the browse view live.
type ResourceMessage struct {
	Kind string `json:"kind"` // always "resource"
	ID   string `json:"id"`
	On   bool   `json:"on"`
}

// EventMessage announces a newly recorded history row for a watched
// resource — same shape as the GET /api/events response.
type EventMessage struct {
	Kind         string `json:"kind"` // always "event"
	ID           int64  `json:"id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	On           bool   `json:"on"`
	Outcome      string `json:"outcome"`
	CreatedAt    string `json:"created_at"`
}

// BridgeStatusMessage announces a transition in the eventstream connection
// to the Hue bridge itself, so clients don't have to infer it from a stale
// initial snapshot.
type BridgeStatusMessage struct {
	Kind   string `json:"kind"` // always "bridge_status"
	Online bool   `json:"online"`
}
