package stable

// Config holds parameters for stability detection.
type Config struct {
	WindowSize int
	Eps        float64
}

// NewFromConfig builds a window from config values.
func NewFromConfig(cfg Config) *Window {
	return NewWindow(cfg.WindowSize, cfg.Eps)
}

// DefaultConfig returns spec-aligned defaults.
func DefaultConfig() Config {
	return Config{
		WindowSize: 10,
		Eps:        5.0,
	}
}

// Snapshot captures window state for diagnostics.
type Snapshot struct {
	Count  int     `json:"count"`
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"stdDev"`
	Stable bool    `json:"stable"`
	Eps    float64 `json:"eps"`
}

// SnapshotWindow returns diagnostic snapshot of window state.
func SnapshotWindow(w *Window) Snapshot {
	if w == nil {
		return Snapshot{}
	}
	return Snapshot{
		Count:  w.Count(),
		Mean:   w.Mean(),
		StdDev: w.StdDev(),
		Stable: w.Stable(),
		Eps:    w.Eps(),
	}
}
