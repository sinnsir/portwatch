package notifier_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/runner"
)

func newTestNotifier(t *testing.T) *notifier.Notifier {
	t.Helper()
	r := runner.New(log.New(os.Stderr, "", 0))
	return notifier.New(r, log.New(os.Stderr, "", 0))
}

func TestNotifier_WebhookFired(t *testing.T) {
	fired := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fired = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := newTestNotifier(t)
	entry := config.PortEntry{
		Port: 8080,
		Host: "localhost",
		Actions: []config.Action{{On: "open", Webhook: ts.URL}},
	}
	n.Notify(entry, "open")
	if !fired {
		t.Error("expected webhook to be fired")
	}
}

func TestNotifier_ActionSkippedWrongState(t *testing.T) {
	fired := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fired = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := newTestNotifier(t)
	entry := config.PortEntry{
		Port: 8080,
		Host: "localhost",
		Actions: []config.Action{{On: "closed", Webhook: ts.URL}},
	}
	n.Notify(entry, "open")
	if fired {
		t.Error("webhook should not fire when state does not match")
	}
}

func TestNotifier_CommandFired(t *testing.T) {
	n := newTestNotifier(t)
	entry := config.PortEntry{
		Port: 9090,
		Host: "localhost",
		Actions: []config.Action{{On: "closed", Command: "echo port-closed"}},
	}
	// Should not panic or error
	n.Notify(entry, "closed")
}
