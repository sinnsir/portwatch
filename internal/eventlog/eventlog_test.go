package eventlog

import (
	"testing"
	"time"
)

func makeEntry(port int, state string) Entry {
	return Entry{
		Host:  "localhost",
		Port:  port,
		State: state,
	}
}

func TestAppend_StoresEntry(t *testing.T) {
	l := New(Policy{})
	l.Append(makeEntry(8080, "open"))
	if l.Len() != 1 {
		t.Fatalf("expected 1 entry, got %d", l.Len())
	}
}

func TestAppend_TimestampAutoSet(t *testing.T) {
	l := New(Policy{})
	l.Append(makeEntry(8080, "open"))
	snap := l.Snapshot()
	if snap[0].Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestAppend_EvictsOldestWhenFull(t *testing.T) {
	l := New(Policy{MaxEntries: 3})
	for i := 1; i <= 4; i++ {
		l.Append(makeEntry(i, "open"))
	}
	snap := l.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(snap))
	}
	if snap[0].Port != 2 {
		t.Fatalf("expected oldest port=2, got %d", snap[0].Port)
	}
}

func TestSnapshot_ReturnsCopy(t *testing.T) {
	l := New(Policy{})
	l.Append(makeEntry(9000, "closed"))
	a := l.Snapshot()
	a[0].State = "mutated"
	b := l.Snapshot()
	if b[0].State == "mutated" {
		t.Fatal("snapshot is not a copy")
	}
}

func TestClear_RemovesAllEntries(t *testing.T) {
	l := New(Policy{})
	l.Append(makeEntry(80, "open"))
	l.Clear()
	if l.Len() != 0 {
		t.Fatalf("expected 0 entries after Clear, got %d", l.Len())
	}
}

func TestDefaultMaxEntries_UsedOnZeroPolicy(t *testing.T) {
	l := New(Policy{MaxEntries: 0})
	for i := 0; i < DefaultMaxEntries+10; i++ {
		l.Append(makeEntry(i, "open"))
	}
	if l.Len() != DefaultMaxEntries {
		t.Fatalf("expected %d, got %d", DefaultMaxEntries, l.Len())
	}
}

func TestQuery_FilterByPort(t *testing.T) {
	l := New(Policy{})
	l.Append(makeEntry(80, "open"))
	l.Append(makeEntry(443, "open"))
	res := l.Query(Filter{Port: 80})
	if len(res) != 1 || res[0].Port != 80 {
		t.Fatalf("unexpected query result: %+v", res)
	}
}

func TestQuery_FilterByState(t *testing.T) {
	l := New(Policy{})
	l.Append(makeEntry(80, "open"))
	l.Append(makeEntry(81, "closed"))
	res := l.Query(Filter{State: "closed"})
	if len(res) != 1 || res[0].State != "closed" {
		t.Fatalf("unexpected query result: %+v", res)
	}
}

func TestQuery_FilterBySince(t *testing.T) {
	l := New(Policy{})
	old := Entry{Host: "localhost", Port: 80, State: "open", Timestamp: time.Now().Add(-2 * time.Hour)}
	recent := Entry{Host: "localhost", Port: 80, State: "open", Timestamp: time.Now()}
	l.Append(old)
	l.Append(recent)
	res := l.Query(Filter{Since: time.Now().Add(-1 * time.Hour)})
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
}

func TestQuery_Limit(t *testing.T) {
	l := New(Policy{})
	for i := 0; i < 10; i++ {
		l.Append(makeEntry(i, "open"))
	}
	res := l.Query(Filter{Limit: 3})
	if len(res) != 3 {
		t.Fatalf("expected 3, got %d", len(res))
	}
}
