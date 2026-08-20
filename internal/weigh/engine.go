package weigh

import (
	"sync"
	"time"

	"github.com/lacsar712/dockload/internal/calib"
)

// Engine coordinates multiple hook processors and optional callbacks.
type Engine struct {
	registry  *Registry
	onPublish func(*Event)
	mu        sync.RWMutex
}

// NewEngine wraps a registry with optional publish callback.
func NewEngine(reg *Registry) *Engine {
	return &Engine{registry: reg}
}

// OnPublish registers callback invoked when events are produced.
func (e *Engine) OnPublish(fn func(*Event)) {
	e.mu.Lock()
	e.onPublish = fn
	e.mu.Unlock()
}

// IngestSample routes sample to hook processor.
func (e *Engine) IngestSample(hookID string, raw int64, cal CalibrationView, ts time.Time) Result {
	proc := e.registry.Get(hookID)
	res := proc.Process(raw, cal, ts)
	if res.Event != nil {
		e.mu.RLock()
		fn := e.onPublish
		e.mu.RUnlock()
		if fn != nil {
			fn(res.Event)
		}
	}
	return res
}

// HookCount returns number of known hooks.
func (e *Engine) HookCount() int {
	e.registry.mu.RLock()
	defer e.registry.mu.RUnlock()
	return len(e.registry.processors)
}

// ResetAll clears every hook FSM.
func (e *Engine) ResetAll() {
	e.registry.mu.Lock()
	defer e.registry.mu.Unlock()
	for _, p := range e.registry.processors {
		p.Reset()
	}
}

// Diagnostics returns FSM snapshots for monitoring.
func (e *Engine) Diagnostics() []Snapshot {
	return e.registry.Snapshots()
}

// RunStableSequence feeds identical samples until publish or max attempts.
func RunStableSequence(proc *HookProcessor, cal CalibrationView, raw int64, n int, ts time.Time) *Event {
	var last Result
	for i := 0; i < n; i++ {
		last = proc.Process(raw, cal, ts.Add(time.Duration(i)*time.Millisecond))
		if last.Event != nil {
			return last.Event
		}
	}
	return last.Event
}

// FixedCalibration is an in-memory CalibrationView for tests.
type FixedCalibration struct {
	C calibCalibration
}

type calibCalibration struct {
	Tare      int64
	Span      float64
	MaxLoadKg float64
}

// Apply implements CalibrationView using fixed values.
func (f FixedCalibration) Apply(raw int64) (float64, error) {
	if f.C.Span <= 0 {
		return 0, calib.ErrNotCalibrated
	}
	net := float64(raw+f.C.Tare) * f.C.Span
	if f.C.MaxLoadKg > 0 && (net > f.C.MaxLoadKg || net < -f.C.MaxLoadKg) {
		return net, calib.ErrOverRange
	}
	return net, nil
}

// NewFixedCalibration builds test calibration view.
func NewFixedCalibration(tare int64, span, maxLoad float64) FixedCalibration {
	return FixedCalibration{C: calibCalibration{Tare: tare, Span: span, MaxLoadKg: maxLoad}}
}
