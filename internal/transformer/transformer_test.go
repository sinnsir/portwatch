package transformer

import (
	"strings"
	"testing"
	"time"
)

func baseEvent() *Event {
	return &Event{
		Host:  "localhost",
		Port:  8080,
		State: "open",
	}
}

func TestApply_AllStepsRun(t *testing.T) {
	called := 0
	step := func(e *Event) error { called++; return nil }
	tr := New(step, step, step)
	if err := tr.Apply(baseEvent()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 3 {
		t.Fatalf("expected 3 calls, got %d", called)
	}
}

func TestApply_StopsOnError(t *testing.T) {
	called := 0
	fail := func(e *Event) error { return fmt.Errorf("boom") }
	after := func(e *Event) error { called++; return nil }
	tr := New(fail, after)
	if err := tr.Apply(baseEvent()); err == nil {
		t.Fatal("expected error, got nil")
	}
	if called != 0 {
		t.Fatalf("step after error should not run, got %d calls", called)
	}
}

func TestLen(t *testing.T) {
	tr := New(NormaliseHost, EnsureTimestamp)
	if tr.Len() != 2 {
		t.Fatalf("expected 2, got %d", tr.Len())
	}
}

func TestNormaliseHost(t *testing.T) {
	e := &Event{Host: "  EXAMPLE.COM  ", Port: 80, State: "open"}
	if err := NormaliseHost(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Host != "example.com" {
		t.Fatalf("expected example.com, got %q", e.Host)
	}
}

func TestEnsureTimestamp_SetsWhenZero(t *testing.T) {
	e := baseEvent()
	before := time.Now()
	if err := EnsureTimestamp(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Timestamp.Before(before) {
		t.Fatal("timestamp should be at or after before")
	}
}

func TestEnsureTimestamp_PreservesExisting(t *testing.T) {
	e := baseEvent()
	fixed := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	e.Timestamp = fixed
	_ = EnsureTimestamp(e)
	if !e.Timestamp.Equal(fixed) {
		t.Fatal("existing timestamp should not be overwritten")
	}
}

func TestAddMeta_MergesValues(t *testing.T) {
	e := baseEvent()
	fn := AddMeta(map[string]string{"env": "prod", "region": "us-east-1"})
	if err := fn(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Meta["env"] != "prod" || e.Meta["region"] != "us-east-1" {
		t.Fatalf("meta not merged correctly: %v", e.Meta)
	}
}

func TestRequirePort_Valid(t *testing.T) {
	for _, p := range []int{1, 80, 65535} {
		e := &Event{Host: "h", Port: p, State: "open"}
		if err := RequirePort(e); err != nil {
			t.Fatalf("port %d should be valid, got %v", p, err)
		}
	}
}

func TestRequirePort_Invalid(t *testing.T) {
	for _, p := range []int{0, -1, 65536} {
		e := &Event{Host: "h", Port: p, State: "open"}
		if err := RequirePort(e); err == nil {
			t.Fatalf("port %d should be invalid", p)
		} else if !strings.Contains(err.Error(), "invalid port") {
			t.Fatalf("unexpected error text: %v", err)
		}
	}
}
