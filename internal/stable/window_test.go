package stable_test

import (
	"testing"

	"github.com/lacsar712/dockload/internal/stable"
)

func TestWindowStable(t *testing.T) {
	w := stable.NewWindow(5, 0.5)
	for i := 0; i < 5; i++ {
		w.Push(1000)
	}
	if !w.Stable() {
		t.Fatalf("expected stable, stddev=%f", w.StdDev())
	}
}

func TestWindowNotStableUntilFull(t *testing.T) {
	w := stable.NewWindow(5, 0.5)
	w.Push(1000)
	if w.Stable() {
		t.Fatal("should not be stable with partial window")
	}
}

func TestPopulationStdDev(t *testing.T) {
	vals := []int64{10, 12, 14, 16, 18}
	sd := stable.PopulationStdDev(vals)
	if sd < 2.8 || sd > 3.0 {
		t.Fatalf("unexpected stddev %f", sd)
	}
}

func TestSamplesOrder(t *testing.T) {
	w := stable.NewWindow(3, 1)
	w.Push(1)
	w.Push(2)
	w.Push(3)
	s := w.Samples()
	if len(s) != 3 || s[0] != 1 || s[2] != 3 {
		t.Fatalf("bad samples: %v", s)
	}
}

func TestWindowReset(t *testing.T) {
	w := stable.NewWindow(3, 1)
	w.Push(1)
	w.Push(2)
	w.Reset()
	if w.Count() != 0 {
		t.Fatal("expected empty after reset")
	}
}
