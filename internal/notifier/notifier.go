package notifier

import (
	"fmt"
	"log"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/runner"
)

// Notifier dispatches actions when a port changes state.
type Notifier struct {
	runner *runner.Runner
	logger *log.Logger
}

// New creates a Notifier with the given runner and logger.
func New(r *runner.Runner, logger *log.Logger) *Notifier {
	return &Notifier{runner: r, logger: logger}
}

// Notify fires all actions configured for the given state transition.
func (n *Notifier) Notify(entry config.PortEntry, state string) {
	for _, action := range entry.Actions {
		if action.On != state {
			continue
		}
		payload := buildPayload(entry, state)
		if action.Webhook != "" {
			if err := n.runner.RunWebhook(action.Webhook, payload); err != nil {
				n.logger.Printf("[notifier] webhook error for port %d: %v", entry.Port, err)
			} else {
				n.logger.Printf("[notifier] webhook sent for port %d state=%s", entry.Port, state)
			}
		}
		if action.Command != "" {
			if err := n.runner.RunCommand(action.Command, payload); err != nil {
				n.logger.Printf("[notifier] command error for port %d: %v", entry.Port, err)
			} else {
				n.logger.Printf("[notifier] command executed for port %d state=%s", entry.Port, state)
			}
		}
	}
}

// buildPayload constructs a human-readable payload string for actions.
func buildPayload(entry config.PortEntry, state string) string {
	return fmt.Sprintf(`{"port":%d,"host":%q,"state":%q}`, entry.Port, entry.Host, state)
}
