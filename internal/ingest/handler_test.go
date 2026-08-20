package ingest_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lacsar712/dockload/internal/calib"
	"github.com/lacsar712/dockload/internal/ingest"
	"github.com/lacsar712/dockload/internal/publish"
	"github.com/lacsar712/dockload/internal/weigh"
)

func setupPipeline(t *testing.T) (*ingest.Processor, *publish.MemoryPublisher) {
	t.Helper()
	store := calib.NewStore()
	_ = store.Put("H7", calib.Calibration{Tare: 1000, Span: 0.05, MaxLoadKg: 45000})
	pub := publish.NewMemoryPublisher(10)
	reg := weigh.NewRegistry(5, 0.5, 50)
	status := weigh.NewStatusRegistry()
	proc := ingest.NewProcessor(reg, store, pub, status)
	return proc, pub
}

func TestHandleRawPublish(t *testing.T) {
	proc, pub := setupPipeline(t)
	ctx := context.Background()
	ts := time.Now()
	readings := ingest.SimulateStableSequence("H7", 1500, 5, ts)
	var last weigh.IngestResponse
	for _, r := range readings {
		last = proc.HandleRaw(ctx, r)
	}
	_ = last
	if pub.Count() != 1 {
		t.Fatalf("expected 1 published event, got %d", pub.Count())
	}
}

func TestValidatorHookID(t *testing.T) {
	v := ingest.DefaultValidator()
	if err := v.ValidateHookID("H7"); err != nil {
		t.Fatal(err)
	}
	if err := v.ValidateHookID(""); err == nil {
		t.Fatal("expected error for empty hook")
	}
}

func TestNormalizeHookID(t *testing.T) {
	if ingest.NormalizeHookID(" h7 ") != "H7" {
		t.Fatal("normalize failed")
	}
}

func TestDecodeReading(t *testing.T) {
	body := strings.NewReader(`{"hookID":"H7","rawCounts":1500}`)
	var dec ingest.Decoder
	r, err := dec.Decode(body)
	if err != nil {
		t.Fatal(err)
	}
	if r.RawCounts != 1500 {
		t.Fatalf("bad raw %d", r.RawCounts)
	}
}
