// Package weigh implements the per-hook weighing finite state machine.
package weigh

import (
	"sync"
	"time"

	"github.com/lacsar712/dockload/internal/calib"
	"github.com/lacsar712/dockload/internal/stable"
)

// State represents FSM phase for a single hook.
type State int

const (
	StateIdle State = iota
	StateCollecting
	StateStable
	StatePublished
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateCollecting:
		return "Collecting"
	case StateStable:
		return "Stable"
	case StatePublished:
		return "Published"
	default:
		return "Unknown"
	}
}

// RejectReason enumerates why a sample did not produce an event.
type RejectReason string

const (
	RejectNone           RejectReason = ""
	RejectNotCalibrated  RejectReason = "NOT_CALIBRATED"
	RejectOverRange      RejectReason = "OVER_RANGE"
	RejectNotStable      RejectReason = "NOT_STABLE"
	RejectDuplicate      RejectReason = "DUPLICATE"
)

// Event is a published weigh result for TOS consumption.
type Event struct {
	HookID    string    `json:"hookID"`
	NetKg     float64   `json:"netKg"`
	RawCounts int64     `json:"rawCounts"`
	MeanRaw   float64   `json:"meanRaw"`
	StdDev    float64   `json:"stdDev"`
	Timestamp time.Time `json:"timestamp"`
}

// Result captures the outcome of processing one raw sample.
type Result struct {
	Event        *Event
	Reject       RejectReason
	State        State
	PreviousState State
}

// HookProcessor runs the weigh FSM for one hook.
type HookProcessor struct {
	hookID     string
	window     *stable.Window
	loadChange float64
	state      State
	published  bool
}

// NewHookProcessor creates an FSM instance for hookID.
func NewHookProcessor(hookID string, window *stable.Window, loadChange float64) *HookProcessor {
	return &HookProcessor{
		hookID:     hookID,
		window:     window,
		loadChange: loadChange,
		state:      StateIdle,
	}
}

// State returns current FSM state.
func (p *HookProcessor) State() State {
	return p.state
}

// HookID returns the hook identifier.
func (p *HookProcessor) HookID() string {
	return p.hookID
}

// Reset clears window and returns FSM to Idle.
func (p *HookProcessor) Reset() {
	p.window.Reset()
	p.state = StateIdle
	p.published = false
}

// Process ingests one raw sample through calibration and FSM logic.
//
// FSM: Idle → Collecting → Stable → Published → Idle
// From Stable, if load jumps beyond loadChange threshold, return to Collecting.
func (p *HookProcessor) Process(raw int64, cal CalibrationView, ts time.Time) Result {
	prev := p.state
	res := Result{PreviousState: prev}

	// Guard: an uncalibrated or out-of-range hook must never reach the
	// publish path. A zero span or missing calibration surfaces as
	// ErrNotCalibrated; reject and block the event so downstream billing
	// never consumes uncalibrated weight.
	netKg, err := cal.Apply(raw)
	if err != nil {
		res.Reject = mapCalibError(err)
		res.State = p.state
		return res
	}

	switch p.state {
	case StateIdle:
		p.window.Reset()
		p.published = false
		p.state = StateCollecting

	case StateStable, StatePublished:
		if last, ok := p.window.Last(); ok {
			if absFloat(float64(raw)-float64(last)) > p.loadChange {
				p.window.Reset()
				p.published = false
				p.state = StateCollecting
			}
		}
	}

	p.window.Push(raw)

	switch p.state {
	case StateCollecting:
		if p.window.Stable() {
			p.state = StateStable
		}
	case StateStable, StatePublished:
		if !p.window.Stable() {
			p.state = StateCollecting
			p.published = false
		}
	}

	if (p.state == StateStable || p.state == StatePublished) && p.window.Stable() && !p.published {
		ev := &Event{
			HookID:    p.hookID,
			NetKg:     netKg,
			RawCounts: raw,
			MeanRaw:   p.window.Mean(),
			StdDev:    p.window.StdDev(),
			Timestamp: ts,
		}
		res.Event = ev
		p.published = true
		p.state = StatePublished
	}

	if res.Event == nil && res.Reject == RejectNone {
		res.Reject = RejectNotStable
	}

	res.State = p.state
	return res
}

// CalibrationView abstracts calibration lookup for testing.
type CalibrationView interface {
	Apply(raw int64) (float64, error)
}

// StoreCalibrationView wraps calib.Store for a fixed hook.
type StoreCalibrationView struct {
	Store  *calib.Store
	HookID string
}

// Apply implements CalibrationView.
func (v StoreCalibrationView) Apply(raw int64) (float64, error) {
	return calib.ApplyForHook(v.Store, v.HookID, raw)
}

func mapCalibError(err error) RejectReason {
	switch {
	case calib.IsNotCalibrated(err):
		return RejectNotCalibrated
	case calib.IsOverRange(err):
		return RejectOverRange
	default:
		return RejectNotCalibrated
	}
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Snapshot describes FSM diagnostics.
type Snapshot struct {
	HookID string          `json:"hookID"`
	State  string          `json:"state"`
	Window stable.Snapshot `json:"window"`
}

// SnapshotProcessor returns diagnostic view.
func SnapshotProcessor(p *HookProcessor) Snapshot {
	return Snapshot{
		HookID: p.hookID,
		State:  p.state.String(),
		Window: stable.SnapshotWindow(p.window),
	}
}

// Registry holds processors for all hooks.
type Registry struct {
	mu          sync.RWMutex
	processors  map[string]*HookProcessor
	windowSize  int
	stableEps   float64
	loadChange  float64
}

// NewRegistry creates hook processor registry.
func NewRegistry(windowSize int, stableEps, loadChange float64) *Registry {
	return &Registry{
		processors: make(map[string]*HookProcessor),
		windowSize: windowSize,
		stableEps:  stableEps,
		loadChange: loadChange,
	}
}

// Get returns processor for hook, creating if needed.
func (r *Registry) Get(hookID string) *HookProcessor {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.processors[hookID]
	if !ok {
		w := stable.NewWindow(r.windowSize, r.stableEps)
		p = NewHookProcessor(hookID, w, r.loadChange)
		r.processors[hookID] = p
	}
	return p
}

// ResetHook clears one hook's FSM.
func (r *Registry) ResetHook(hookID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.processors[hookID]; ok {
		p.Reset()
	}
}

// Snapshots returns diagnostics for all hooks.
func (r *Registry) Snapshots() []Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Snapshot, 0, len(r.processors))
	for _, p := range r.processors {
		out = append(out, SnapshotProcessor(p))
	}
	return out
}
