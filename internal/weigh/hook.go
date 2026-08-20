package weigh

import (
	"sync"
)

// HookStatus aggregates runtime status for one hook.
type HookStatus struct {
	HookID string  `json:"hookID"`
	State  string  `json:"state"`
	Stable bool    `json:"stable"`
	Mean   float64 `json:"meanRaw"`
	StdDev float64 `json:"stdDev"`
}

// StatusRegistry tracks live hook statuses for the web UI.
type StatusRegistry struct {
	mu       sync.RWMutex
	statuses map[string]HookStatus
}

// NewStatusRegistry creates an empty status registry.
func NewStatusRegistry() *StatusRegistry {
	return &StatusRegistry{statuses: make(map[string]HookStatus)}
}

// Update records latest processor snapshot.
func (sr *StatusRegistry) Update(p *HookProcessor) {
	if p == nil {
		return
	}
	snap := SnapshotProcessor(p)
	st := HookStatus{
		HookID: snap.HookID,
		State:  snap.State,
		Stable: snap.Window.Stable,
		Mean:   snap.Window.Mean,
		StdDev: snap.Window.StdDev,
	}
	sr.mu.Lock()
	sr.statuses[p.HookID()] = st
	sr.mu.Unlock()
}

// List returns all hook statuses.
func (sr *StatusRegistry) List() []HookStatus {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	out := make([]HookStatus, 0, len(sr.statuses))
	for _, st := range sr.statuses {
		out = append(out, st)
	}
	return out
}

// StatusResponse wraps hook statuses for API.
type StatusResponse struct {
	Items []HookStatus `json:"items"`
}

// BuildStatusResponse converts list to API shape.
func BuildStatusResponse(items []HookStatus) StatusResponse {
	return StatusResponse{Items: items}
}
