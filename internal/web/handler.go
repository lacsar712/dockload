package web

import (
	"encoding/json"
	"net/http"
)

// DashboardData aggregates UI polling payload.
type DashboardData struct {
	Calibrations []CalibRow   `json:"calibrations"`
	RecentWeighs []WeighRow   `json:"recentWeighs"`
	HookStatus   []StatusRow  `json:"hookStatus"`
}

// CalibRow is one calibration table row.
type CalibRow struct {
	HookID    string  `json:"hookID"`
	Tare      int64   `json:"tare"`
	Span      float64 `json:"span"`
	MaxLoadKg float64 `json:"maxLoadKg"`
}

// WeighRow is one recent weigh row.
type WeighRow struct {
	HookID    string  `json:"hookID"`
	NetKg     float64 `json:"netKg"`
	RawCounts int64   `json:"rawCounts"`
	Timestamp string  `json:"timestamp"`
}

// StatusRow is live hook status for UI.
type StatusRow struct {
	HookID string  `json:"hookID"`
	State  string  `json:"state"`
	Stable bool    `json:"stable"`
	Mean   float64 `json:"meanRaw"`
	StdDev float64 `json:"stdDev"`
}

// WriteDashboard encodes dashboard JSON to response.
func WriteDashboard(w http.ResponseWriter, data DashboardData) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

// IndexHTML returns the main page filename in embed FS.
const IndexHTML = "index.html"

// ContentTypeFor returns MIME type for static asset names.
func ContentTypeFor(name string) string {
	switch {
	case hasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case hasSuffix(name, ".css"):
		return "text/css; charset=utf-8"
	case hasSuffix(name, ".js"):
		return "application/javascript; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func hasSuffix(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}
