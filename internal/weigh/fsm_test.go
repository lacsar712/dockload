package weigh_test

import (
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/weigh"
)

func TestFSMPublishStableLoad(t *testing.T) {
	w := weigh.NewRegistry(5, 0.5, 50)
	proc := w.Get("H7")
	cal := weigh.NewFixedCalibration(1000, 0.05, 45000)
	ts := time.Now()

	var ev *weigh.Event
	for i := 0; i < 5; i++ {
		res := proc.Process(1500, cal, ts)
		if res.Event != nil {
			ev = res.Event
		}
	}
	if ev == nil {
		t.Fatal("expected published event")
	}
	if ev.NetKg != 25 {
		t.Fatalf("expected 25 kg, got %f", ev.NetKg)
	}
}

func TestFSMNotCalibrated(t *testing.T) {
	proc := weigh.NewHookProcessor("H1", nil, 50)
	w := weigh.NewRegistry(3, 1, 50)
	proc = w.Get("H1")
	cal := weigh.NewFixedCalibration(0, 0, 0)
	res := proc.Process(100, cal, time.Now())
	if res.Reject != weigh.RejectNotCalibrated {
		t.Fatalf("expected NOT_CALIBRATED, got %s", res.Reject)
	}
}

func TestFSMOverRange(t *testing.T) {
	w := weigh.NewRegistry(3, 1, 50)
	proc := w.Get("H1")
	cal := weigh.NewFixedCalibration(0, 1, 100)
	res := proc.Process(200, cal, time.Now())
	if res.Reject != weigh.RejectOverRange {
		t.Fatalf("expected OVER_RANGE, got %s", res.Reject)
	}
}

func TestRunStableSequence(t *testing.T) {
	w := weigh.NewRegistry(4, 0.1, 50)
	proc := w.Get("H7")
	cal := weigh.NewFixedCalibration(1000, 0.05, 45000)
	ev := weigh.RunStableSequence(proc, cal, 1200, 10, time.Now())
	if ev == nil {
		t.Fatal("expected event from sequence")
	}
}
