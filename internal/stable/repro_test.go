package stable_test

import (
	"math"
	"testing"

	"github.com/lacsar712/dockload/internal/stable"
)

// TestWindowStableUsesPopulationStdDev reproduces dockload-002:
// Window.StdDev must be population (n), not sample (n-1).
func TestWindowStableUsesPopulationStdDev(t *testing.T) {
	// population sd ≈ 2.828; sample sd ≈ 3.162
	w := stable.NewWindow(5, 3.0)
	for _, v := range []int64{10, 12, 14, 16, 18} {
		w.Push(v)
	}
	sd := w.StdDev()
	pop := stable.PopulationStdDev([]int64{10, 12, 14, 16, 18})
	if math.Abs(sd-pop) > 1e-9 {
		t.Fatalf("Window.StdDev=%f want population %f (sample would be ~3.16)", sd, pop)
	}
	if !w.Stable() {
		t.Fatalf("expected stable with population stddev=%f <= eps=3", sd)
	}
}

// TestStableInclusiveEps reproduces dockload-005: Stable when StdDev == eps.
func TestStableInclusiveEps(t *testing.T) {
	// n=2 values {0,2}: population stddev = 1
	w := stable.NewWindow(2, 1.0)
	w.Push(0)
	w.Push(2)
	if w.StdDev() != 1 {
		t.Fatalf("stddev=%f want 1", w.StdDev())
	}
	if !w.Stable() {
		t.Fatal("expected Stable when StdDev == eps (inclusive <=)")
	}
	if !stable.IsStable([]int64{0, 2}, 2, 1.0) {
		t.Fatal("expected IsStable when PopulationStdDev == eps")
	}
}

// TestSamplesChronologicalAfterWrap reproduces dockload-009.
func TestSamplesChronologicalAfterWrap(t *testing.T) {
	w := stable.NewWindow(3, 1)
	w.Push(10)
	w.Push(20)
	w.Push(30)
	w.Push(40) // evicts 10; chronological: 20,30,40
	s := w.Samples()
	if len(s) != 3 || s[0] != 20 || s[1] != 30 || s[2] != 40 {
		t.Fatalf("Samples after wrap want [20 30 40], got %v", s)
	}
}
