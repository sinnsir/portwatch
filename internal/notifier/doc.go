// Package notifier provides state-change notification dispatch for portwatch.
//
// When the monitor detects that a port has transitioned between open and closed
// states, the Notifier evaluates all configured actions for the affected port
// and fires webhooks or shell commands whose "on" field matches the new state.
//
// Example usage:
//
//	r := runner.New(logger)
//	n := notifier.New(r, logger)
//	n.Notify(entry, "open")
package notifier
