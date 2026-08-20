package app_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lacsar712/dockload/internal/app"
	"github.com/lacsar712/dockload/internal/config"
)

func newTestApp(t *testing.T) *app.App {
	t.Helper()
	cfg := config.Default()
	cfg.Addr = ":0"
	a, err := app.New(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestHealthEndpoint(t *testing.T) {
	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestPutAndGetCalib(t *testing.T) {
	a := newTestApp(t)
	body := bytes.NewBufferString(`{"tare":1000,"span":0.05,"maxLoadKg":45000}`)
	req := httptest.NewRequest(http.MethodPut, "/v1/calib/H7", body)
	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put status %d body %s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/calib/H7", nil)
	rec2 := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("get status %d", rec2.Code)
	}
}

func TestRawIngestFlow(t *testing.T) {
	a := newTestApp(t)
	putBody := bytes.NewBufferString(`{"tare":1000,"span":0.05,"maxLoadKg":45000}`)
	putReq := httptest.NewRequest(http.MethodPut, "/v1/calib/H7", putBody)
	a.Handler().ServeHTTP(httptest.NewRecorder(), putReq)

	for i := 0; i < 10; i++ {
		payload, _ := json.Marshal(map[string]any{
			"hookID": "H7", "rawCounts": 1500,
		})
		req := httptest.NewRequest(http.MethodPost, "/v1/sensors/raw", bytes.NewReader(payload))
		a.Handler().ServeHTTP(httptest.NewRecorder(), req)
	}

	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/weighs/recent", nil))
	var resp struct {
		Count int `json:"count"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Count < 1 {
		t.Fatalf("expected recent weighs, got %d", resp.Count)
	}
}

func TestIndexPage(t *testing.T) {
	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("index status %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("dockload")) {
		t.Fatal("expected index html content")
	}
}
