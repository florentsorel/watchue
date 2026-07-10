package handler

import "sync"

// NotifierStore holds the currently active Notifier (if any), safe for
// concurrent use between the HTTP handler and the event-loop goroutine.
//
// Unlike the Hue client — tied to a long-lived SSE-subscription goroutine,
// hence reconfigured via a full process restart — a Notifier makes only
// one-shot outbound calls, so it can be swapped in place: saving a new
// provider via the /setup notification step takes effect on the very next
// event, no restart needed.
type NotifierStore struct {
	mu       sync.RWMutex
	notifier Notifier
}

// NewNotifierStore returns an empty store (Get returns nil until Set is called).
func NewNotifierStore() *NotifierStore {
	return &NotifierStore{}
}

func (s *NotifierStore) Get() Notifier {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.notifier
}

func (s *NotifierStore) Set(n Notifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifier = n
}
