// Package publish delivers WeighEvent instances to TOS and internal buffers.
package publish

import (
	"context"
	"sync"

	"github.com/lacsar712/dockload/internal/weigh"
)

// Publisher sends weigh events to downstream consumers.
type Publisher interface {
	Publish(ctx context.Context, event *weigh.Event) error
}

// MemoryPublisher retains recent events in a ring buffer.
type MemoryPublisher struct {
	mu      sync.RWMutex
	max     int
	events  []*weigh.Event
	subs    []chan *weigh.Event
	closed  bool
}

// NewMemoryPublisher creates an in-memory publisher with capacity max.
func NewMemoryPublisher(max int) *MemoryPublisher {
	if max < 1 {
		max = 1
	}
	return &MemoryPublisher{max: max, events: make([]*weigh.Event, 0, max)}
}

// Publish stores event and notifies subscribers.
func (m *MemoryPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if event == nil {
		return nil
	}

	cp := weigh.CloneEvent(event)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return context.Canceled
	}

	m.events = append(m.events, cp)
	if len(m.events) > m.max {
		m.events = m.events[len(m.events)-m.max:]
	}

	for _, ch := range m.subs {
		select {
		case ch <- cp:
		default:
		}
	}
	return nil
}

// Recent returns newest-first copy of stored events.
func (m *MemoryPublisher) Recent(limit int) []*weigh.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := len(m.events)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]*weigh.Event, limit)
	for i := 0; i < limit; i++ {
		src := m.events[n-1-i]
		out[i] = weigh.CloneEvent(src)
	}
	return out
}

// Subscribe registers a channel for live events. Caller must drain or unsubscribe.
func (m *MemoryPublisher) Subscribe(buf int) (<-chan *weigh.Event, func()) {
	ch := make(chan *weigh.Event, buf)
	m.mu.Lock()
	m.subs = append(m.subs, ch)
	m.mu.Unlock()

	unsub := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		for i, sub := range m.subs {
			if sub == ch {
				m.subs = append(m.subs[:i], m.subs[i+1:]...)
				close(ch)
				break
			}
		}
	}
	return ch, unsub
}

// Close shuts down publisher and subscribers.
func (m *MemoryPublisher) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	for _, ch := range m.subs {
		close(ch)
	}
	m.subs = nil
}

// Count returns number of stored events.
func (m *MemoryPublisher) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.events)
}
