package weigh

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventJSON is the wire format for weigh events.
type EventJSON struct {
	HookID    string  `json:"hookID"`
	NetKg     float64 `json:"netKg"`
	RawCounts int64   `json:"rawCounts"`
	MeanRaw   float64 `json:"meanRaw"`
	StdDev    float64 `json:"stdDev"`
	Timestamp string  `json:"timestamp"`
}

// ToJSON converts Event to API representation.
func (e *Event) ToJSON() EventJSON {
	return EventJSON{
		HookID:    e.HookID,
		NetKg:     e.NetKg,
		RawCounts: e.RawCounts,
		MeanRaw:   e.MeanRaw,
		StdDev:    e.StdDev,
		Timestamp: e.Timestamp.UTC().Format(time.RFC3339),
	}
}

// EncodeEvent writes event as JSON line.
func EncodeEvent(e *Event, w json.Encoder) error {
	if e == nil {
		return fmt.Errorf("nil event")
	}
	return w.Encode(e.ToJSON())
}

// RecentResponse wraps recent events for GET /v1/weighs/recent.
type RecentResponse struct {
	Items []EventJSON `json:"items"`
	Count int         `json:"count"`
}

// BuildRecentResponse converts events to API response.
func BuildRecentResponse(events []*Event) RecentResponse {
	items := make([]EventJSON, 0, len(events))
	for _, e := range events {
		if e != nil {
			items = append(items, e.ToJSON())
		}
	}
	return RecentResponse{Items: items, Count: len(items)}
}

// CloneEvent returns a deep copy of event.
func CloneEvent(e *Event) *Event {
	if e == nil {
		return nil
	}
	cp := *e
	return &cp
}
