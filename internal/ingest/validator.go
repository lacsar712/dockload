package ingest

import (
	"fmt"
	"strings"
	"unicode"
)

// Validator provides extended validation rules for sensor input.
type Validator struct {
	MaxHookIDLen int
	MaxRaw       int64
}

// DefaultValidator returns production validation limits.
func DefaultValidator() Validator {
	return Validator{
		MaxHookIDLen: 64,
		MaxRaw:       1_000_000_000,
	}
}

// ValidateHookID checks hook identifier format.
func (v Validator) ValidateHookID(hookID string) error {
	if hookID == "" {
		return fmt.Errorf("hookID is required")
	}
	if len(hookID) > v.MaxHookIDLen {
		return fmt.Errorf("hookID exceeds max length %d", v.MaxHookIDLen)
	}
	for _, r := range hookID {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return fmt.Errorf("hookID contains invalid character %q", r)
		}
	}
	return nil
}

// ValidateRaw checks raw count bounds.
func (v Validator) ValidateRaw(raw int64) error {
	if raw < 0 {
		return fmt.Errorf("rawCounts must be non-negative")
	}
	if raw > v.MaxRaw {
		return fmt.Errorf("rawCounts exceeds maximum %d", v.MaxRaw)
	}
	return nil
}

// ValidateReading applies all validator rules.
func (v Validator) ValidateReading(r RawReading) error {
	if err := v.ValidateHookID(r.HookID); err != nil {
		return err
	}
	if err := v.ValidateRaw(r.RawCounts); err != nil {
		return err
	}
	return nil
}

// NormalizeHookID trims and uppercases hook IDs for consistency.
func NormalizeHookID(hookID string) string {
	return strings.ToUpper(strings.TrimSpace(hookID))
}
