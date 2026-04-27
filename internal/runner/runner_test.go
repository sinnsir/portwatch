package runner_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"portwatch/internal/runner"
)

func testPayload() runner.Payload {
	return runner.Payload{
		Port:  8080,
		Host:  "localhost",
		State: "open",
		Time:  time.Now().UTC().Format(time.RFC3339),
	}
}

func TestRunWebhook_Success(t *testing.T) {
	var received runner.Payload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := runner.New()
	p := testPayload()
	if err := r.RunWebhook(ts.URL, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Port != p.Port || received.State != p.State {
		t.Errorf("payload mismatch: got %+v, want %+v", received, p)
	}
}

func TestRunWebhook_Non2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	r := runner.New()
	if err := r.RunWebhook(ts.URL, testPayload()); err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestRunCommand_Success(t *testing.T) {
	r := runner.New()
	p := testPayload()
	// Simple command that uses injected env vars and exits 0.
	if err := r.RunCommand("test -n \"$PW_PORT\"", p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCommand_Failure(t *testing.T) {
	r := runner.New()
	if err := r.RunCommand("exit 1", testPayload()); err == nil {
		t.Fatal("expected error for failing command, got nil")
	}
}
