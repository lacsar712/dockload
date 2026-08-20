package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/lacsar712/dockload/internal/weigh"
)

// HTTPPublisher POSTs weigh events to a TOS webhook endpoint.
type HTTPPublisher struct {
	client  *http.Client
	url     string
	headers map[string]string
	logger  *log.Logger
}

// HTTPPublisherConfig configures outbound webhook delivery.
type HTTPPublisherConfig struct {
	URL     string
	Timeout time.Duration
	Headers map[string]string
	Logger  *log.Logger
}

// NewHTTPPublisher creates a webhook publisher.
func NewHTTPPublisher(cfg HTTPPublisherConfig) *HTTPPublisher {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPPublisher{
		client:  &http.Client{Timeout: timeout},
		url:     cfg.URL,
		headers: cfg.Headers,
		logger:  cfg.Logger,
	}
}

// Publish sends event to configured URL.
func (h *HTTPPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	if h.url == "" || event == nil {
		return nil
	}
	body, err := json.Marshal(event.ToJSON())
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("post event: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("tos returned status %d", resp.StatusCode)
	}
	if h.logger != nil {
		h.logger.Printf("published weigh hook=%s net=%.2f", event.HookID, event.NetKg)
	}
	return nil
}

// MultiPublisher fans out to multiple publishers.
type MultiPublisher struct {
	targets []Publisher
}

// NewMultiPublisher combines publishers.
func NewMultiPublisher(targets ...Publisher) *MultiPublisher {
	return &MultiPublisher{targets: targets}
}

// Publish sends to all targets; first error is returned after all attempts.
func (m *MultiPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	var firstErr error
	for _, t := range m.targets {
		if t == nil {
			continue
		}
		if err := t.Publish(ctx, event); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// LoggingPublisher wraps another publisher with structured logs.
type LoggingPublisher struct {
	inner  Publisher
	logger *log.Logger
}

// NewLoggingPublisher decorates publisher with logging.
func NewLoggingPublisher(inner Publisher, logger *log.Logger) *LoggingPublisher {
	return &LoggingPublisher{inner: inner, logger: logger}
}

// Publish logs then delegates.
func (l *LoggingPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	err := l.inner.Publish(ctx, event)
	if l.logger != nil && event != nil {
		if err != nil {
			l.logger.Printf("publish failed hook=%s err=%v", event.HookID, err)
		} else {
			l.logger.Printf("publish ok hook=%s net=%.2f kg", event.HookID, event.NetKg)
		}
	}
	return err
}

// AsyncPublisher queues publishes on a worker goroutine.
type AsyncPublisher struct {
	inner Publisher
	ch    chan asyncJob
	wg    sync.WaitGroup
}

type asyncJob struct {
	ctx   context.Context
	event *weigh.Event
}

// NewAsyncPublisher starts background worker.
func NewAsyncPublisher(inner Publisher, queue int) *AsyncPublisher {
	if queue < 1 {
		queue = 64
	}
	a := &AsyncPublisher{inner: inner, ch: make(chan asyncJob, queue)}
	a.wg.Add(1)
	go a.worker()
	return a
}

func (a *AsyncPublisher) worker() {
	defer a.wg.Done()
	for job := range a.ch {
		_ = a.inner.Publish(job.ctx, job.event)
	}
}

// Publish enqueues event respecting context cancellation.
func (a *AsyncPublisher) Publish(ctx context.Context, event *weigh.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case a.ch <- asyncJob{ctx: ctx, event: event}:
		return nil
	}
}

// Close stops worker after draining.
func (a *AsyncPublisher) Close() {
	close(a.ch)
	a.wg.Wait()
}
