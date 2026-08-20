package app

import (
	"log"
	"net/http"
	"runtime"
	"time"
)

func (a *App) withMiddleware(next http.Handler) http.Handler {
	h := next
	h = a.corsMiddleware(h)
	h = a.logMiddleware(h)
	h = a.recoverMiddleware(h)
	return h
}

func (a *App) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				a.logger.Printf("panic: %v", rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *App) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)
		a.logger.Printf("%s %s %d %s", r.Method, r.URL.Path, lw.status, time.Since(start))
	})
}

func (a *App) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.status = code
	lw.ResponseWriter.WriteHeader(code)
}

// HealthSnapshot describes runtime health for monitoring.
type HealthSnapshot struct {
	Status    string  `json:"status"`
	UptimeSec float64 `json:"uptimeSec"`
	GoVersion string  `json:"goVersion"`
}

// SnapshotHealth returns current health metrics.
func SnapshotHealth(since time.Time) HealthSnapshot {
	return HealthSnapshot{
		Status:    "ok",
		UptimeSec: time.Since(since).Seconds(),
		GoVersion: runtime.Version(),
	}
}

// DefaultLogger creates a prefixed logger for the service.
func DefaultLogger() *log.Logger {
	return log.New(log.Writer(), "[dockload] ", log.LstdFlags|log.Lmsgprefix)
}
