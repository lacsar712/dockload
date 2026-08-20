package publish

import (
	"context"

	"github.com/lacsar712/dockload/internal/weigh"
)

// NopPublisher discards all events (useful in tests).
type NopPublisher struct{}

// Publish implements Publisher.
func (NopPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	return nil
}

// FilterPublisher only publishes events matching predicate.
type FilterPublisher struct {
	inner     Publisher
	predicate func(*weigh.Event) bool
}

// NewFilterPublisher creates conditional publisher.
func NewFilterPublisher(inner Publisher, predicate func(*weigh.Event) bool) *FilterPublisher {
	return &FilterPublisher{inner: inner, predicate: predicate}
}

// Publish applies filter before delegating.
func (f *FilterPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	if f.predicate != nil && !f.predicate(event) {
		return nil
	}
	if f.inner == nil {
		return nil
	}
	return f.inner.Publish(ctx, event)
}

// MinNetFilter returns predicate requiring minimum net kg.
func MinNetFilter(minKg float64) func(*weigh.Event) bool {
	return func(e *weigh.Event) bool {
		return e != nil && e.NetKg >= minKg
	}
}

// HookFilter returns predicate matching hook ID.
func HookFilter(hookID string) func(*weigh.Event) bool {
	return func(e *weigh.Event) bool {
		return e != nil && e.HookID == hookID
	}
}
