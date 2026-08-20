package publish_test

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/publish"
	"github.com/lacsar712/dockload/internal/weigh"
)

// TestHTTPPublishHonorsCancel reproduces dockload-004.
func TestHTTPPublishHonorsCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	block := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
		w.WriteHeader(http.StatusOK)
	})}
	go srv.Serve(ln)
	defer func() {
		close(block)
		_ = srv.Close()
	}()

	pub := publish.NewHTTPPublisher(publish.HTTPPublisherConfig{
		URL:     "http://" + ln.Addr().String() + "/hook",
		Timeout: 5 * time.Second,
	})
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- pub.Publish(ctx, &weigh.Event{HookID: "H7", NetKg: 25, Timestamp: time.Now()})
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error after cancel")
		}
		if !strings.Contains(err.Error(), "context") && err != context.Canceled {
			t.Fatalf("expected cancel-related error, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Publish hung after cancel (likely used context.Background)")
	}
}
