package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/portwatch/internal/api"
	"github.com/user/portwatch/internal/history"
)

// stubHistory satisfies api.HistoryStore.
type stubHistory struct {
	events []history.Event
}

func (s *stubHistory) Snapshot() []history.Event { return s.events }

func makeEvents(n int) []history.Event {
	evts := make([]history.Event, n)
	for i := range evts {
		evts[i] = history.Event{
			Port:      8080,
			Timestamp: time.Now(),
		}
	}
	return evts
}

func TestHandleHealth(t *testing.T) {
	srv := api.New(":0", &stubHistory{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("unexpected status: %s", body["status"])
	}
}

func TestHandleEvents_ReturnsAll(t *testing.T) {
	h := &stubHistory{events: makeEvents(5)}
	srv := api.New(":0", h)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []history.Event
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 5 {
		t.Errorf("expected 5 events, got %d", len(events))
	}
}

func TestHandleEvents_LimitParam(t *testing.T) {
	h := &stubHistory{events: makeEvents(10)}
	srv := api.New(":0", h)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events?limit=3", nil)
	srv.ServeHTTP(rec, req)

	var events []history.Event
	if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestHandleEvents_EmptyHistory(t *testing.T) {
	srv := api.New(":0", &stubHistory{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
