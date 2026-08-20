package weigh

import "encoding/json"

// IngestResponse is returned from POST /v1/sensors/raw.
type IngestResponse struct {
	Accepted bool         `json:"accepted"`
	Reject   RejectReason `json:"reject,omitempty"`
	State    string       `json:"state"`
	Event    *EventJSON   `json:"event,omitempty"`
}

// NewIngestResponse builds API response from processing result.
func NewIngestResponse(res Result) IngestResponse {
	out := IngestResponse{
		Accepted: res.Event != nil,
		Reject:   res.Reject,
		State:    res.State.String(),
	}
	if res.Event != nil {
		j := res.Event.ToJSON()
		out.Event = &j
	}
	return out
}

// RejectResponse documents a rejected weigh attempt.
type RejectResponse struct {
	Reason RejectReason `json:"reason"`
	Detail string       `json:"detail,omitempty"`
}

// EncodeReject writes reject payload.
func EncodeReject(w json.Encoder, reason RejectReason, detail string) error {
	return w.Encode(RejectResponse{Reason: reason, Detail: detail})
}

// ReasonString returns human-readable reject explanation.
func ReasonString(r RejectReason) string {
	switch r {
	case RejectNotCalibrated:
		return "Hook is not calibrated or span is invalid"
	case RejectOverRange:
		return "Net weight exceeds configured maximum load"
	case RejectNotStable:
		return "Load has not stabilized within the window"
	case RejectDuplicate:
		return "Duplicate publish suppressed"
	default:
		return ""
	}
}
