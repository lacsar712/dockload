package publish

import (
	"context"
	"sync"

	"github.com/lacsar712/dockload/internal/weigh"
)

// Subscriber receives published events.
type Subscriber interface {
	OnEvent(ctx context.Context, event *weigh.Event) error
}

// Fanout manages registered subscribers notified on each publish.
type Fanout struct {
	mu   sync.RWMutex
	subs []Subscriber
}

// NewFanout creates empty fanout registry.
func NewFanout() *Fanout {
	return &Fanout{}
}

// Register adds subscriber.
func (f *Fanout) Register(sub Subscriber) {
	f.mu.Lock()
	f.subs = append(f.subs, sub)
	f.mu.Unlock()
}

// Publish notifies all subscribers.
func (f *Fanout) Publish(ctx context.Context, event *weigh.Event) error {
	f.mu.RLock()
	subs := append([]Subscriber(nil), f.subs...)
	f.mu.RUnlock()
	var first error
	for _, s := range subs {
		if err := s.OnEvent(ctx, event); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// CallbackSubscriber adapts function to Subscriber.
type CallbackSubscriber struct {
	Fn func(context.Context, *weigh.Event) error
}

// OnEvent implements Subscriber.
func (c CallbackSubscriber) OnEvent(ctx context.Context, event *weigh.Event) error {
	if c.Fn == nil {
		return nil
	}
	return c.Fn(ctx, event)
}

// BufferedSubscriber collects events in memory.
type BufferedSubscriber struct {
	mu     sync.Mutex
	events []*weigh.Event
}

// NewBufferedSubscriber creates collector.
func NewBufferedSubscriber() *BufferedSubscriber {
	return &BufferedSubscriber{}
}

// OnEvent appends event copy.
func (b *BufferedSubscriber) OnEvent(ctx context.Context, event *weigh.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	b.mu.Lock()
	b.events = append(b.events, weigh.CloneEvent(event))
	b.mu.Unlock()
	return nil
}

// Events returns collected copies.
func (b *BufferedSubscriber) Events() []*weigh.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]*weigh.Event, len(b.events))
	for i, e := range b.events {
		out[i] = weigh.CloneEvent(e)
	}
	return out
}
