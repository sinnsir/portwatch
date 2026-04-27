package history_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/history"
	"github.com/user/portwatch/internal/portcheck"
)

func makeEvent(port int) history.Event {
	return history.Event{
		Port:     port,
		Host:     "localhost",
		OldState: portcheck.StateClosed,
		NewState: portcheck.StateOpen,
	}
}

func TestHistory_RecordAndSnapshot(t *testing.T) {
	h := history.New(10)

	h.Record(makeEvent(8080))
	h.Record(makeEvent(9090))

	snap := h.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 events, got %d", len(snap))
	}
	if snap[0].Port != 8080 {
		t.Errorf("expected first event port 8080, got %d", snap[0].Port)
	}
	if snap[1].Port != 9090 {
		t.Errorf("expected second event port 9090, got %d", snap[1].Port)
	}
}

func TestHistory_RingOverwrite(t *testing.T) {
	h := history.New(3)

	for i := 1; i <= 5; i++ {
		h.Record(makeEvent(i))
	}

	// Only the last 3 events should be retained.
	snap := h.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 events after overflow, got %d", len(snap))
	}
	if snap[0].Port != 3 || snap[1].Port != 4 || snap[2].Port != 5 {
		t.Errorf("unexpected ports in snapshot: %v", snap)
	}
}

func TestHistory_TimestampAutoSet(t *testing.T) {
	h := history.New(5)
	before := time.Now()
	h.Record(makeEvent(1234))
	after := time.Now()

	snap := h.Snapshot()
	ts := snap[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not within expected range [%v, %v]", ts, before, after)
	}
}

func TestHistory_DefaultCapacity(t *testing.T) {
	h := history.New(0)
	for i := 0; i < 150; i++ {
		h.Record(makeEvent(i))
	}
	if h.Len() != 100 {
		t.Errorf("expected default capacity 100, got %d", h.Len())
	}
}

func TestHistory_ConcurrentAccess(t *testing.T) {
	h := history.New(50)
	done := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			h.Record(makeEvent(i))
		}
		close(done)
	}()

	for {
		select {
		case <-done:
			return
		default:
			_ = h.Snapshot()
		}
	}
}
