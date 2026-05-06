package envelope_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/envelope"
)

type samplePayload struct {
	Host  string `json:"host"`
	Port  int    `json:"port"`
	State string `json:"state"`
}

func TestWrap_DefaultFields(t *testing.T) {
	p := samplePayload{Host: "localhost", Port: 8080, State: "open"}
	before := time.Now().UTC()

	env, err := envelope.Wrap(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Version != "v1" {
		t.Errorf("expected version v1, got %q", env.Version)
	}
	if env.Source != "portwatch" {
		t.Errorf("expected source portwatch, got %q", env.Source)
	}
	if env.SentAt.Before(before) {
		t.Error("SentAt should be >= time before Wrap call")
	}
	if len(env.Payload) == 0 {
		t.Error("Payload must not be empty")
	}
}

func TestWrap_WithOptions(t *testing.T) {
	p := samplePayload{Host: "example.com", Port: 443, State: "closed"}

	env, err := envelope.Wrap(p,
		envelope.WithVersion("v2"),
		envelope.WithSource("test-suite"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if env.Version != "v2" {
		t.Errorf("expected v2, got %q", env.Version)
	}
	if env.Source != "test-suite" {
		t.Errorf("expected test-suite, got %q", env.Source)
	}
}

func TestWrap_Unwrap_RoundTrip(t *testing.T) {
	orig := samplePayload{Host: "127.0.0.1", Port: 9090, State: "open"}

	env, err := envelope.Wrap(orig)
	if err != nil {
		t.Fatalf("wrap error: %v", err)
	}

	var got samplePayload
	if err := env.Unwrap(&got); err != nil {
		t.Fatalf("unwrap error: %v", err)
	}

	if got != orig {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, orig)
	}
}

func TestWrap_MarshalJSON_ContainsFields(t *testing.T) {
	p := samplePayload{Host: "h", Port: 1, State: "open"}

	env, _ := envelope.Wrap(p)
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	s := string(b)
	for _, field := range []string{"version", "source", "sent_at", "payload"} {
		if !strings.Contains(s, field) {
			t.Errorf("marshalled JSON missing field %q: %s", field, s)
		}
	}
}

func TestWrap_UnmarshalablePayload(t *testing.T) {
	// channels cannot be marshalled to JSON.
	_, err := envelope.Wrap(make(chan int))
	if err == nil {
		t.Error("expected error for non-serialisable payload")
	}
}
