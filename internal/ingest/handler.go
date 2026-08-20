package ingest

import (
	"context"
	"time"

	"github.com/lacsar712/dockload/internal/calib"
	"github.com/lacsar712/dockload/internal/publish"
	"github.com/lacsar712/dockload/internal/weigh"
)

// Processor wires ingest pipeline: validate → FSM → publish.
type Processor struct {
	Registry   *weigh.Registry
	CalibStore *calib.Store
	Publisher  publish.Publisher
	Status     *weigh.StatusRegistry
	Validator  Validator
	Decoder    Decoder
}

// NewProcessor creates an ingest processor with defaults.
func NewProcessor(reg *weigh.Registry, store *calib.Store, pub publish.Publisher, status *weigh.StatusRegistry) *Processor {
	return &Processor{
		Registry:   reg,
		CalibStore: store,
		Publisher:  pub,
		Status:     status,
		Validator:  DefaultValidator(),
	}
}

// HandleRaw processes a single raw reading through the full pipeline.
func (p *Processor) HandleRaw(ctx context.Context, reading RawReading) weigh.IngestResponse {
	hookID := NormalizeHookID(reading.HookID)
	reading.HookID = hookID

	if err := p.Validator.ValidateReading(reading); err != nil {
		return weigh.IngestResponse{
			Accepted: false,
			Reject:   weigh.RejectNotStable,
			State:    weigh.StateIdle.String(),
		}
	}

	ts := reading.TimestampOrNow()
	proc := p.Registry.Get(hookID)
	view := weigh.StoreCalibrationView{Store: p.CalibStore, HookID: hookID}
	res := proc.Process(reading.RawCounts, view, ts)

	if p.Status != nil {
		p.Status.Update(proc)
	}

	if res.Event != nil && p.Publisher != nil {
		_ = p.Publisher.Publish(ctx, res.Event)
	}

	return weigh.NewIngestResponse(res)
}

// HandleBatch processes multiple readings sequentially.
func (p *Processor) HandleBatch(ctx context.Context, batch BatchReading) []weigh.IngestResponse {
	out := make([]weigh.IngestResponse, 0, len(batch.Items))
	for _, item := range batch.Items {
		out = append(out, p.HandleRaw(ctx, item))
	}
	return out
}

// SimulateStableSequence generates N identical readings for test helpers.
func SimulateStableSequence(hookID string, raw int64, n int, ts time.Time) []RawReading {
	items := make([]RawReading, n)
	for i := 0; i < n; i++ {
		items[i] = RawReading{
			HookID:    hookID,
			RawCounts: raw,
			Timestamp: ts.Add(time.Duration(i) * time.Millisecond).Format(time.RFC3339),
		}
	}
	return items
}
