package calib

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONCodec serializes calibrations for HTTP handlers.
type JSONCodec struct{}

// Decode reads a calibration from JSON body.
func (JSONCodec) Decode(r io.Reader) (Calibration, error) {
	var c Calibration
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Calibration{}, fmt.Errorf("decode calibration: %w", err)
	}
	return c, nil
}

// Encode writes calibration as JSON.
func (JSONCodec) Encode(w io.Writer, c Calibration) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}

// HookResponse is the API shape for GET /v1/calib/{hookID}.
type HookResponse struct {
	HookID string      `json:"hookID"`
	Calib  Calibration `json:"calib"`
}

// ListResponse wraps all hook calibrations.
type ListResponse struct {
	Items []HookResponse `json:"items"`
}

// BuildListResponse converts store snapshot to API response.
func BuildListResponse(store *Store) ListResponse {
	snap := store.List()
	items := make([]HookResponse, 0, len(snap))
	for id, c := range snap {
		items = append(items, HookResponse{HookID: id, Calib: c})
	}
	return ListResponse{Items: items}
}
