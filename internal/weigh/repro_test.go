package weigh_test

import (
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/weigh"
)

// TestFSMNetUsesTareSubtract reproduces dockload-001 via FixedCalibration.
func TestFSMNetUsesTareSubtract(t *testing.T) {
	reg := weigh.NewRegistry(3, 1, 50)
	proc := reg.Get("H7")
	cal := weigh.NewFixedCalibration(1000, 0.05, 45000)
	ts := time.Now()
	var ev *weigh.Event
	for i := 0; i < 3; i++ {
		res := proc.Process(1500, cal, ts)
		if res.Event != nil {
			ev = res.Event
		}
	}
	if ev == nil {
		t.Fatal("expected event")
	}
	if ev.NetKg != 25 {
		t.Fatalf("expected net 25 kg from (1500-1000)*0.05, got %f", ev.NetKg)
	}
}

// TestFSMNegativeOverRange reproduces dockload-010 via FixedCalibration.
func TestFSMNegativeOverRange(t *testing.T) {
	reg := weigh.NewRegistry(3, 1, 50)
	proc := reg.Get("H1")
	cal := weigh.NewFixedCalibration(0, 1, 100)
	res := proc.Process(-200, cal, time.Now())
	if res.Reject != weigh.RejectOverRange {
		t.Fatalf("expected OVER_RANGE, got %s", res.Reject)
	}
}
