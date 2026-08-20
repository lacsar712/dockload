package app

import (
	"context"
	"fmt"
	"time"
)

// Lifecycle manages startup/shutdown hooks for the App.
type Lifecycle struct {
	onStart    []func(context.Context) error
	onShutdown []func(context.Context) error
}

// NewLifecycle creates empty lifecycle registry.
func NewLifecycle() *Lifecycle {
	return &Lifecycle{}
}

// OnStart registers startup hook.
func (l *Lifecycle) OnStart(fn func(context.Context) error) {
	l.onStart = append(l.onStart, fn)
}

// OnShutdown registers shutdown hook.
func (l *Lifecycle) OnShutdown(fn func(context.Context) error) {
	l.onShutdown = append(l.onShutdown, fn)
}

// RunStart executes all startup hooks.
func (l *Lifecycle) RunStart(ctx context.Context) error {
	for i, fn := range l.onStart {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("start hook %d: %w", i, err)
		}
	}
	return nil
}

// RunShutdown executes all shutdown hooks.
func (l *Lifecycle) RunShutdown(ctx context.Context) error {
	var first error
	for i := len(l.onShutdown) - 1; i >= 0; i-- {
		if err := l.onShutdown[i](ctx); err != nil && first == nil {
			first = fmt.Errorf("shutdown hook %d: %w", i, err)
		}
	}
	return first
}

// DefaultLifecycle builds standard hooks for App.
func DefaultLifecycle(a *App) *Lifecycle {
	lc := NewLifecycle()
	lc.OnStart(func(ctx context.Context) error {
		a.Logger().Printf("lifecycle start: listening %s", a.Config.Addr)
		return nil
	})
	lc.OnShutdown(func(ctx context.Context) error {
		shCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return a.Shutdown(shCtx)
	})
	return lc
}

// RunWithLifecycle starts app with lifecycle hooks.
func RunWithLifecycle(ctx context.Context, a *App, lc *Lifecycle) error {
	if lc == nil {
		lc = DefaultLifecycle(a)
	}
	if err := lc.RunStart(ctx); err != nil {
		return err
	}
	return a.Start(ctx)
}
