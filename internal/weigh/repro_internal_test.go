package weigh

import (
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/stable"
)

// TestLoadChangeComparesRawNotNet reproduces dockload-006.
func TestLoadChangeComparesRawNotNet(t *testing.T) {
	w := stable.NewWindow(3, 10)
	p := NewHookProcessor("H1", w, 50) // threshold in raw counts
	cal := NewFixedCalibration(0, 10, 45000) // span=10 → netKg >> raw delta
	ts := time.Now()
	for i := 0; i < 3; i++ {
		p.Process(100, cal, ts)
	}
	if p.state != StatePublished || !p.published {
		t.Fatalf("setup: want Published, got state=%s published=%v", p.state, p.published)
	}
	// raw delta 10 < 50 → must stay Published; buggy netKg=1100 > 50 resets to Collecting
	res := p.Process(110, cal, ts)
	if res.State == StateCollecting {
		t.Fatal("raw delta 10 must not trigger load-change when threshold is 50 counts")
	}
	if res.State != StatePublished {
		t.Fatalf("want Published after small raw move, got %s", res.State)
	}
}

// TestPublishedFlagSuppressesDuplicate reproduces dockload-008.
func TestPublishedFlagSuppressesDuplicate(t *testing.T) {
	w := stable.NewWindow(3, 1)
	p := NewHookProcessor("H1", w, 1000)
	cal := NewFixedCalibration(0, 1, 45000)
	ts := time.Now()
	var events int
	for i := 0; i < 6; i++ {
		res := p.Process(100, cal, ts.Add(time.Duration(i)*time.Millisecond))
		if res.Event != nil {
			events++
		}
	}
	if events != 1 {
		t.Fatalf("expected exactly 1 publish while load stays put, got %d", events)
	}
	if !p.published {
		t.Fatal("published flag must be set after first event")
	}
}
