package ingest_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/calib"
	"github.com/lacsar712/dockload/internal/ingest"
	"github.com/lacsar712/dockload/internal/publish"
	"github.com/lacsar712/dockload/internal/weigh"
)

// TestHandleRawRejectsUncalibrated reproduces dockload-003.
func TestHandleRawRejectsUncalibrated(t *testing.T) {
	store := calib.NewStore()
	pub := publish.NewMemoryPublisher(10)
	reg := weigh.NewRegistry(3, 1, 50)
	proc := ingest.NewProcessor(reg, store, pub, weigh.NewStatusRegistry())
	ctx := context.Background()
	ts := time.Now()
	var last weigh.IngestResponse
	for _, r := range ingest.SimulateStableSequence("H9", 1500, 5, ts) {
		last = proc.HandleRaw(ctx, r)
	}
	if pub.Count() != 0 {
		t.Fatalf("uncalibrated hook must not publish, got %d events", pub.Count())
	}
	if last.Reject != weigh.RejectNotCalibrated {
		t.Fatalf("expected NOT_CALIBRATED, got %s", last.Reject)
	}
}

// TestHandleRawNormalizesHookID reproduces dockload-007.
func TestHandleRawNormalizesHookID(t *testing.T) {
	store := calib.NewStore()
	if err := store.Put("H7", calib.Calibration{Tare: 1000, Span: 0.05, MaxLoadKg: 45000}); err != nil {
		t.Fatal(err)
	}
	pub := publish.NewMemoryPublisher(10)
	reg := weigh.NewRegistry(3, 0.5, 50)
	proc := ingest.NewProcessor(reg, store, pub, weigh.NewStatusRegistry())
	ctx := context.Background()
	ts := time.Now()
	readings := ingest.SimulateStableSequence(" h7 ", 1500, 5, ts)
	for _, r := range readings {
		_ = proc.HandleRaw(ctx, r)
	}
	if pub.Count() != 1 {
		t.Fatalf("normalized hookID must hit calib H7 and publish once, got %d", pub.Count())
	}
}

// TestHandleRawHonorsCancel reproduces dockload-004 ingest→publish path.
func TestHandleRawHonorsCancel(t *testing.T) {
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

	store := calib.NewStore()
	_ = store.Put("H7", calib.Calibration{Tare: 0, Span: 1, MaxLoadKg: 45000})
	httpPub := publish.NewHTTPPublisher(publish.HTTPPublisherConfig{
		URL:     "http://" + ln.Addr().String() + "/hook",
		Timeout: 5 * time.Second,
	})
	reg := weigh.NewRegistry(3, 1, 50)
	proc := ingest.NewProcessor(reg, store, httpPub, nil)
	ts := time.Now()
	for i := 0; i < 2; i++ {
		_ = proc.HandleRaw(context.Background(), ingest.RawReading{HookID: "H7", RawCounts: 100, Timestamp: ts.Format(time.RFC3339)})
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = proc.HandleRaw(ctx, ingest.RawReading{HookID: "H7", RawCounts: 100, Timestamp: ts.Format(time.RFC3339)})
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("HandleRaw hung after cancel (publish likely used context.Background)")
	}
}
