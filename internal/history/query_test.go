package history

import (
	"testing"
	"time"
)

func TestQuery_FilterByPort(t *testing.T) {
	h := New(20)
	h.Record(Event{Port: 8080, State: "open"})
	h.Record(Event{Port: 9090, State: "open"})
	h.Record(Event{Port: 8080, State: "closed"})

	results := h.Query(Filter{Port: 8080})
	if len(results) != 2 {
		t.Fatalf("expected 2 events for port 8080, got %d", len(results))
	}
	for _, e := range results {
		if e.Port != 8080 {
			t.Errorf("unexpected port %d in results", e.Port)
		}
	}
}

func TestQuery_FilterByState(t *testing.T) {
	h := New(20)
	h.Record(Event{Port: 8080, State: "open"})
	h.Record(Event{Port: 8080, State: "closed"})
	h.Record(Event{Port: 9090, State: "open"})

	results := h.Query(Filter{State: "open"})
	if len(results) != 2 {
		t.Fatalf("expected 2 open events, got %d", len(results))
	}
}

func TestQuery_FilterBySince(t *testing.T) {
	h := New(20)
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now().Add(-1 * time.Minute)

	h.Record(Event{Port: 8080, State: "open", Timestamp: old})
	h.Record(Event{Port: 8080, State: "closed", Timestamp: recent})

	results := h.Query(Filter{Since: time.Now().Add(-30 * time.Minute)})
	if len(results) != 1 {
		t.Fatalf("expected 1 recent event, got %d", len(results))
	}
	if results[0].State != "closed" {
		t.Errorf("expected closed state, got %s", results[0].State)
	}
}

func TestQuery_Limit(t *testing.T) {
	h := New(20)
	for i := 0; i < 10; i++ {
		h.Record(Event{Port: 8080, State: "open"})
	}

	results := h.Query(Filter{Limit: 3})
	if len(results) != 3 {
		t.Fatalf("expected 3 results with limit, got %d", len(results))
	}
}

func TestQuery_NoFilter_ReturnsAll(t *testing.T) {
	h := New(20)
	h.Record(Event{Port: 8080, State: "open"})
	h.Record(Event{Port: 9090, State: "closed"})

	results := h.Query(Filter{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results with no filter, got %d", len(results))
	}
}

func TestQuery_EmptyHistory(t *testing.T) {
	h := New(10)
	results := h.Query(Filter{Port: 8080})
	if results != nil && len(results) != 0 {
		t.Fatalf("expected empty results for empty history, got %d", len(results))
	}
}
