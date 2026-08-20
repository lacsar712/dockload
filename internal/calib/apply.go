package calib

import "fmt"

// ApplyForHook loads calibration from store and converts raw counts.
func ApplyForHook(store *Store, hookID string, raw int64) (float64, error) {
	if store == nil {
		return 0, fmt.Errorf("calibration store is nil")
	}
	c, ok := store.Get(hookID)
	if !ok || !c.IsSet() {
		return 0, ErrNotCalibrated
	}
	return c.Apply(raw)
}

// EnsureDefaultMaxLoad sets MaxLoadKg when unset on incoming calibration.
func EnsureDefaultMaxLoad(c Calibration, defaultMax float64) Calibration {
	if c.MaxLoadKg <= 0 && defaultMax > 0 {
		c.MaxLoadKg = defaultMax
	}
	return c
}

// FormatNet formats net weight for display with two decimal places.
func FormatNet(kg float64) string {
	return fmt.Sprintf("%.2f", kg)
}

// RawFromNet inverts Apply for test fixtures: raw = net/span + tare.
func RawFromNet(c Calibration, netKg float64) (int64, error) {
	if c.Span <= 0 {
		return 0, ErrNotCalibrated
	}
	raw := int64(netKg/c.Span) + c.Tare
	return raw, nil
}
