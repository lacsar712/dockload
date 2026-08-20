// Package calib manages tare/span calibration and raw-to-kg conversion.
package calib

import (
	"fmt"
	"sync"
)

// Calibration holds per-hook tare and span parameters.
type Calibration struct {
	Tare      int64   `json:"tare"`
	Span      float64 `json:"span"`
	MaxLoadKg float64 `json:"maxLoadKg"`
}

// IsSet reports whether span has been configured for use.
func (c Calibration) IsSet() bool {
	return c.Span > 0
}

// Apply converts raw sensor counts to net kilograms.
// Formula: netKg = (raw - tare) * span
func (c Calibration) Apply(raw int64) (float64, error) {
	if c.Span <= 0 {
		return 0, ErrNotCalibrated
	}
	net := float64(raw-c.Tare) * c.Span
	if c.MaxLoadKg > 0 && (net > c.MaxLoadKg || net < -c.MaxLoadKg) {
		return net, ErrOverRange
	}
	return net, nil
}

// Validate checks calibration fields before persistence.
func (c Calibration) Validate() error {
	if c.Span <= 0 {
		return fmt.Errorf("span must be positive, got %f", c.Span)
	}
	if c.MaxLoadKg < 0 {
		return fmt.Errorf("maxLoadKg must be non-negative, got %f", c.MaxLoadKg)
	}
	return nil
}

// Clone returns a copy safe for external use.
func (c Calibration) Clone() Calibration {
	return c
}

// Store provides concurrent read/write access to hook calibrations.
type Store struct {
	mu    sync.RWMutex
	items map[string]Calibration
}

// NewStore creates an empty calibration store.
func NewStore() *Store {
	return &Store{items: make(map[string]Calibration)}
}

// Get returns calibration for hookID and whether it exists.
func (s *Store) Get(hookID string) (Calibration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.items[hookID]
	return c, ok
}

// Put stores calibration for hookID after validation.
func (s *Store) Put(hookID string, c Calibration) error {
	if hookID == "" {
		return fmt.Errorf("hookID must not be empty")
	}
	if err := c.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[hookID] = c.Clone()
	return nil
}

// Delete removes calibration for hookID.
func (s *Store) Delete(hookID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, hookID)
}

// List returns a snapshot of all calibrations keyed by hook ID.
func (s *Store) List() map[string]Calibration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]Calibration, len(s.items))
	for k, v := range s.items {
		out[k] = v.Clone()
	}
	return out
}

// Count returns number of calibrated hooks.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}
