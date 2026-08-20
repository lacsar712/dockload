package calib_test

import (
	"testing"

	"github.com/lacsar712/dockload/internal/calib"
)

// TestApplySubtractsTare reproduces dockload-001: netKg = (raw - tare) * span.
func TestApplySubtractsTare(t *testing.T) {
	c := calib.Calibration{Tare: 1000, Span: 0.05, MaxLoadKg: 45000}
	net, err := c.Apply(1500)
	if err != nil {
		t.Fatal(err)
	}
	if net != 25 {
		t.Fatalf("expected (1500-1000)*0.05=25, got %f", net)
	}
	// raw+tare would yield 125
	if net == 125 {
		t.Fatal("Apply used raw+tare instead of raw-tare")
	}
}

// TestApplyNegativeOverRange reproduces dockload-010: net below -maxLoadKg is OVER_RANGE.
func TestApplyNegativeOverRange(t *testing.T) {
	c := calib.Calibration{Tare: 0, Span: 1, MaxLoadKg: 100}
	_, err := c.Apply(-200)
	if !calib.IsOverRange(err) {
		t.Fatalf("expected OVER_RANGE for net=-200, got %v", err)
	}
}
