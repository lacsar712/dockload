package publish_test

import (
	"context"
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/publish"
	"github.com/lacsar712/dockload/internal/weigh"
)

func TestMemoryPublisher(t *testing.T) {
	pub := publish.NewMemoryPublisher(5)
	ctx := context.Background()
	ev := &weigh.Event{HookID: "H7", NetKg: 25, Timestamp: time.Now()}
	if err := pub.Publish(ctx, ev); err != nil {
		t.Fatal(err)
	}
	if pub.Count() != 1 {
		t.Fatalf("expected 1 event, got %d", pub.Count())
	}
	recent := pub.Recent(1)
	if len(recent) != 1 || recent[0].NetKg != 25 {
		t.Fatal("recent mismatch")
	}
}

func TestMultiPublisher(t *testing.T) {
	a := publish.NewMemoryPublisher(2)
	b := publish.NewMemoryPublisher(2)
	multi := publish.NewMultiPublisher(a, b)
	ev := &weigh.Event{HookID: "H1", NetKg: 10, Timestamp: time.Now()}
	_ = multi.Publish(context.Background(), ev)
	if a.Count() != 1 || b.Count() != 1 {
		t.Fatal("multi publish failed")
	}
}

func TestContextCancel(t *testing.T) {
	pub := publish.NewMemoryPublisher(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := pub.Publish(ctx, &weigh.Event{HookID: "H1", Timestamp: time.Now()})
	if err == nil {
		t.Fatal("expected context error")
	}
}
