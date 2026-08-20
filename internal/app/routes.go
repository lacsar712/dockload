package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lacsar712/dockload/internal/calib"
	"github.com/lacsar712/dockload/internal/ingest"
	"github.com/lacsar712/dockload/internal/weigh"
	"github.com/lacsar712/dockload/internal/web"
)

var startTime = time.Now()

func (a *App) registerRoutes(mux *http.ServeMux) {
	web.Register(mux)

	mux.HandleFunc("GET /health", a.handleHealth)
	mux.HandleFunc("POST /v1/sensors/raw", a.handleRaw)
	mux.HandleFunc("PUT /v1/calib/{hookID}", a.handlePutCalib)
	mux.HandleFunc("GET /v1/calib/{hookID}", a.handleGetCalib)
	mux.HandleFunc("GET /v1/calib", a.handleListCalib)
	mux.HandleFunc("GET /v1/weighs/recent", a.handleRecentWeighs)
	mux.HandleFunc("GET /v1/hooks/status", a.handleHookStatus)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	payload := map[string]any{
		"status":    "ok",
		"uptimeSec": time.Since(startTime).Seconds(),
		"hooks":     a.Registry.Snapshots(),
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func (a *App) handleRaw(w http.ResponseWriter, r *http.Request) {
	var dec ingest.Decoder
	reading, err := dec.Decode(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp := a.Ingest.HandleRaw(r.Context(), reading)
	status := http.StatusOK
	if !resp.Accepted && resp.Reject != "" {
		status = http.StatusAccepted
	}
	writeJSON(w, status, resp)
}

func (a *App) handlePutCalib(w http.ResponseWriter, r *http.Request) {
	hookID := ingest.NormalizeHookID(r.PathValue("hookID"))
	if hookID == "" {
		writeError(w, http.StatusBadRequest, "hookID required")
		return
	}
	var codec calib.JSONCodec
	c, err := codec.Decode(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	c = calib.EnsureDefaultMaxLoad(c, a.Config.DefaultMaxLoad)
	if err := a.CalibStore.Put(hookID, c); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.Registry.ResetHook(hookID)
	writeJSON(w, http.StatusOK, calib.HookResponse{HookID: hookID, Calib: c})
}

func (a *App) handleGetCalib(w http.ResponseWriter, r *http.Request) {
	hookID := ingest.NormalizeHookID(r.PathValue("hookID"))
	c, ok := a.CalibStore.Get(hookID)
	if !ok {
		writeError(w, http.StatusNotFound, "calibration not found")
		return
	}
	writeJSON(w, http.StatusOK, calib.HookResponse{HookID: hookID, Calib: c})
}

func (a *App) handleListCalib(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, calib.BuildListResponse(a.CalibStore))
}

func (a *App) handleRecentWeighs(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 {
			limit = n
		}
	}
	events := a.Publisher.Recent(limit)
	jsonEvents := make([]*weigh.Event, len(events))
	for i, e := range events {
		jsonEvents[i] = e
	}
	writeJSON(w, http.StatusOK, weigh.BuildRecentResponse(jsonEvents))
}

func (a *App) handleHookStatus(w http.ResponseWriter, r *http.Request) {
	items := a.Status.List()
	writeJSON(w, http.StatusOK, weigh.BuildStatusResponse(items))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ParseHookIDFromPath extracts hook ID from legacy path patterns.
func ParseHookIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}
