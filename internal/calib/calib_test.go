package calib_test

import (
	"testing"

	"github.com/lacsar712/dockload/internal/calib"
)

func TestApplyNetWeight(t *testing.T) {
	c := calib.Calibration{Tare: 1000, Span: 0.05, MaxLoadKg: 45000}
	net, err := c.Apply(1500)
	if err != nil {
		t.Fatal(err)
	}
	if net != 25 {
		t.Fatalf("expected 25 kg, got %f", net)
	}
}

func TestApplyNotCalibrated(t *testing.T) {
	c := calib.Calibration{Tare: 1000, Span: 0}
	_, err := c.Apply(1500)
	if !calib.IsNotCalibrated(err) {
		t.Fatalf("expected NOT_CALIBRATED, got %v", err)
	}
}

func TestApplyOverRange(t *testing.T) {
	c := calib.Calibration{Tare: 0, Span: 1, MaxLoadKg: 100}
	_, err := c.Apply(200)
	if !calib.IsOverRange(err) {
		t.Fatalf("expected OVER_RANGE, got %v", err)
	}
}

func TestStoreConcurrent(t *testing.T) {
	store := calib.NewStore()
	c := calib.Calibration{Tare: 1000, Span: 0.05, MaxLoadKg: 45000}
	if err := store.Put("H7", c); err != nil {
		t.Fatal(err)
	}
	got, ok := store.Get("H7")
	if !ok || got.Tare != 1000 {
		t.Fatalf("unexpected calib: %+v ok=%v", got, ok)
	}
}

func TestRawFromNet(t *testing.T) {
	c := calib.Calibration{Tare: 1000, Span: 0.05, MaxLoadKg: 45000}
	raw, err := calib.RawFromNet(c, 25)
	if err != nil {
		t.Fatal(err)
	}
	if raw != 1500 {
		t.Fatalf("expected raw 1500, got %d", raw)
	}
}
