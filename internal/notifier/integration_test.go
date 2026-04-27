package notifier_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/runner"
)

func TestNotifier_MultipleActionsForSameState(t *testing.T) {
	var count int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	r := runner.New(log.New(os.Stderr, "", 0))
	n := notifier.New(r, log.New(os.Stderr, "", 0))

	entry := config.PortEntry{
		Port: 7070,
		Host: "localhost",
		Actions: []config.Action{
			{On: "open", Webhook: ts.URL},
			{On: "open", Webhook: ts.URL},
			{On: "closed", Webhook: ts.URL},
		},
	}
	n.Notify(entry, "open")

	if got := atomic.LoadInt32(&count); got != 2 {
		t.Errorf("expected 2 webhook calls, got %d", got)
	}
}

func TestNotifier_NoActionsNoError(t *testing.T) {
	r := runner.New(log.New(os.Stderr, "", 0))
	n := notifier.New(r, log.New(os.Stderr, "", 0))

	entry := config.PortEntry{
		Port: 5050,
		Host: "localhost",
		Actions: []config.Action{},
	}
	// Must not panic
	n.Notify(entry, "closed")
}
