package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/portwatch/internal/metrics"
)

func newTestRegistry(t *testing.T) *metrics.Registry {
	t.Helper()
	reg := metrics.New()
	reg.Counter("port_checks_total").Add(5)
	reg.Gauge("open_ports").Set(3)
	return reg
}

func TestMetricsHandler_ReturnsSnapshot(t *testing.T) {
	reg := newTestRegistry(t)
	h := newMetricsHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var result map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result["port_checks_total"] != 5 {
		t.Errorf("expected port_checks_total=5, got %d", result["port_checks_total"])
	}
	if result["open_ports"] != 3 {
		t.Errorf("expected open_ports=3, got %d", result["open_ports"])
	}
}

func TestMetricsHandler_MethodNotAllowed(t *testing.T) {
	reg := metrics.New()
	h := newMetricsHandler(reg)

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		req := httptest.NewRequest(method, "/metrics", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, rec.Code)
		}
	}
}

func TestMetricsHandler_EmptyRegistry(t *testing.T) {
	reg := metrics.New()
	h := newMetricsHandler(reg)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}
