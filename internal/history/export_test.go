package history

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"portwatch/internal/portcheck"
)

func TestExportJSON_ValidOutput(t *testing.T) {
	h := New(8)
	h.Record(makeEvent("localhost", 8080, portcheck.StateOpen))
	h.Record(makeEvent("localhost", 9090, portcheck.StateClosed))

	var buf bytes.Buffer
	if err := h.ExportJSON(&buf); err != nil {
		t.Fatalf("ExportJSON returned error: %v", err)
	}

	var events []Event
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Port != 8080 {
		t.Errorf("expected first port 8080, got %d", events[0].Port)
	}
	if events[1].State != portcheck.StateClosed {
		t.Errorf("expected second state closed, got %s", events[1].State)
	}
}

func TestExportTable_ContainsHeaders(t *testing.T) {
	h := New(4)
	ts := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	e := makeEvent("example.com", 443, portcheck.StateOpen)
	e.Timestamp = ts
	h.Record(e)

	var buf bytes.Buffer
	if err := h.ExportTable(&buf); err != nil {
		t.Fatalf("ExportTable returned error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"TIME", "HOST", "PORT", "STATE", "example.com", "443", "open"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestExportJSON_EmptyHistory(t *testing.T) {
	h := New(4)
	var buf bytes.Buffer
	if err := h.ExportJSON(&buf); err != nil {
		t.Fatalf("ExportJSON on empty history returned error: %v", err)
	}
	var events []Event
	if err := json.Unmarshal(buf.Bytes(), &events); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}
