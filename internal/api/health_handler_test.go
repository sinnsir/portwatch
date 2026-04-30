package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/portwatch/internal/api"
	"github.com/yourorg/portwatch/internal/healthcheck"
)

func TestHealthHandler_Healthy(t *testing.T) {
	probe := healthcheck.Probe{Name: "store", Fn: func() error { return nil }}
	checker := healthcheck.New(probe)

	mux := http.NewServeMux()
	api.RegisterHealthHandler(mux, "/healthz", checker)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var status healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !status.Healthy {
		t.Error("expected healthy=true in response")
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	probe := healthcheck.Probe{Name: "db", Fn: func() error { return errors.New("timeout") }}
	checker := healthcheck.New(probe)

	mux := http.NewServeMux()
	api.RegisterHealthHandler(mux, "/healthz", checker)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var status healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if status.Healthy {
		t.Error("expected healthy=false in response")
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	checker := healthcheck.New()

	mux := http.NewServeMux()
	api.RegisterHealthHandler(mux, "/healthz", checker)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
