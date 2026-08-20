package calib

import "errors"

var (
	// ErrNotCalibrated indicates span is zero or hook has no calibration.
	ErrNotCalibrated = errors.New("NOT_CALIBRATED")
	// ErrOverRange indicates computed net weight exceeds maxLoadKg.
	ErrOverRange = errors.New("OVER_RANGE")
)

// RejectReason maps calibration errors to API-facing reason strings.
func RejectReason(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ErrNotCalibrated):
		return "NOT_CALIBRATED"
	case errors.Is(err, ErrOverRange):
		return "OVER_RANGE"
	default:
		return "CALIB_ERROR"
	}
}

// IsNotCalibrated reports whether err is a missing-calibration error.
func IsNotCalibrated(err error) bool {
	return errors.Is(err, ErrNotCalibrated)
}

// IsOverRange reports whether err is an over-range error.
func IsOverRange(err error) bool {
	return errors.Is(err, ErrOverRange)
}
