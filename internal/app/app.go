// Package app wires HTTP server, routing, and dependency injection.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/lacsar712/dockload/internal/calib"
	"github.com/lacsar712/dockload/internal/config"
	"github.com/lacsar712/dockload/internal/ingest"
	"github.com/lacsar712/dockload/internal/publish"
	"github.com/lacsar712/dockload/internal/weigh"
)

// App is the root application container.
type App struct {
	Config     config.Config
	CalibStore *calib.Store
	Registry   *weigh.Registry
	Publisher  *publish.MemoryPublisher
	Status     *weigh.StatusRegistry
	Ingest     *ingest.Processor
	Server     *http.Server
	logger     *log.Logger
	mu         sync.Mutex
	started    bool
}

// New creates an App from configuration.
func New(cfg config.Config, logger *log.Logger) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	if logger == nil {
		logger = log.Default()
	}

	store := calib.NewStore()
	pub := publish.NewMemoryPublisher(cfg.MaxRecent)
	reg := weigh.NewRegistry(cfg.WindowSize, cfg.StableEps, cfg.LoadChange)
	status := weigh.NewStatusRegistry()
	proc := ingest.NewProcessor(reg, store, pub, status)

	a := &App{
		Config:     cfg,
		CalibStore: store,
		Registry:   reg,
		Publisher:  pub,
		Status:     status,
		Ingest:     proc,
		logger:     logger,
	}

	mux := http.NewServeMux()
	a.registerRoutes(mux)

	a.Server = &http.Server{
		Addr:         cfg.Addr,
		Handler:      a.withMiddleware(mux),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return a, nil
}

// Logger returns app logger.
func (a *App) Logger() *log.Logger {
	return a.logger
}

// Start begins listening; blocks until context cancelled or error.
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.started {
		a.mu.Unlock()
		return fmt.Errorf("app already started")
	}
	a.started = true
	a.mu.Unlock()

	a.logger.Printf("dockload starting on %s (%s)", a.Config.Addr, a.Config.String())

	errCh := make(chan error, 1)
	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.Config.ShutdownGrace)
		defer cancel()
		if err := a.Server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		a.Publisher.Close()
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}
}

// Shutdown gracefully stops the server.
func (a *App) Shutdown(ctx context.Context) error {
	a.Publisher.Close()
	return a.Server.Shutdown(ctx)
}

// Handler returns the root HTTP handler for testing.
func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	a.registerRoutes(mux)
	return a.withMiddleware(mux)
}
