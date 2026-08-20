// Package ingest handles sensor raw count submission and validation.
package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// RawReading is the payload from the load sensor gateway.
type RawReading struct {
	HookID    string `json:"hookID"`
	RawCounts int64  `json:"rawCounts"`
	Timestamp string `json:"timestamp,omitempty"`
}

// Validate checks required fields on a raw reading.
func (r RawReading) Validate() error {
	if r.HookID == "" {
		return fmt.Errorf("hookID is required")
	}
	if r.RawCounts < 0 {
		return fmt.Errorf("rawCounts must be non-negative")
	}
	return nil
}

// TimestampOrNow parses RFC3339 timestamp or returns now UTC.
func (r RawReading) TimestampOrNow() time.Time {
	if r.Timestamp == "" {
		return time.Now().UTC()
	}
	t, err := time.Parse(time.RFC3339, r.Timestamp)
	if err != nil {
		return time.Now().UTC()
	}
	return t.UTC()
}

// Decoder reads raw readings from JSON bodies.
type Decoder struct{}

// Decode parses one RawReading from reader.
func (Decoder) Decode(r io.Reader) (RawReading, error) {
	var reading RawReading
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&reading); err != nil {
		return RawReading{}, fmt.Errorf("decode raw reading: %w", err)
	}
	if err := reading.Validate(); err != nil {
		return RawReading{}, err
	}
	return reading, nil
}

// BatchReading supports multiple samples in one POST for bulk gateways.
type BatchReading struct {
	Items []RawReading `json:"items"`
}

// Validate ensures batch has at least one valid item.
func (b BatchReading) Validate() error {
	if len(b.Items) == 0 {
		return fmt.Errorf("batch must contain at least one item")
	}
	for i, item := range b.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("item %d: %w", i, err)
		}
	}
	return nil
}

// DecodeBatch parses batch payload.
func (d Decoder) DecodeBatch(r io.Reader) (BatchReading, error) {
	var batch BatchReading
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&batch); err != nil {
		return BatchReading{}, fmt.Errorf("decode batch: %w", err)
	}
	if err := batch.Validate(); err != nil {
		return BatchReading{}, err
	}
	return batch, nil
}
