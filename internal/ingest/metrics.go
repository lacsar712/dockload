package ingest

import (
	"sync"
	"sync/atomic"
)

// Metrics tracks ingest throughput and outcomes.
type Metrics struct {
	Total    atomic.Uint64
	Accepted atomic.Uint64
	Rejected atomic.Uint64
	ByReason sync.Map
}

// NewMetrics creates zeroed metrics.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordIngest updates counters for one ingest response outcome.
func (m *Metrics) RecordIngest(accepted bool, reject string) {
	m.Total.Add(1)
	if accepted {
		m.Accepted.Add(1)
	} else {
		m.Rejected.Add(1)
		if reject != "" {
			key := reject
			if v, ok := m.ByReason.Load(key); ok {
				m.ByReason.Store(key, v.(uint64)+1)
			} else {
				m.ByReason.Store(key, uint64(1))
			}
		}
	}
}

// Snapshot returns current metric values.
func (m *Metrics) Snapshot() MetricsSnapshot {
	snap := MetricsSnapshot{
		Total:    m.Total.Load(),
		Accepted: m.Accepted.Load(),
		Rejected: m.Rejected.Load(),
		ByReason: make(map[string]uint64),
	}
	m.ByReason.Range(func(k, v any) bool {
		snap.ByReason[k.(string)] = v.(uint64)
		return true
	})
	return snap
}

// MetricsSnapshot is JSON-serializable metrics view.
type MetricsSnapshot struct {
	Total    uint64            `json:"total"`
	Accepted uint64            `json:"accepted"`
	Rejected uint64            `json:"rejected"`
	ByReason map[string]uint64 `json:"byReason"`
}

// AcceptanceRate returns accepted/total ratio.
func (s MetricsSnapshot) AcceptanceRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Accepted) / float64(s.Total)
}
