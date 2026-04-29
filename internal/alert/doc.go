// Package alert implements a rate-limiting filter for port-state change alerts.
//
// It prevents notification spam by suppressing repeated alerts for the same
// (port, state) pair within a configurable cooldown window. This is useful
// when a port flaps rapidly and you only want to be notified once per event
// until the cooldown has elapsed.
//
// Basic usage:
//
//	policy := alert.Policy{Cooldown: 30 * time.Second}
//	f := alert.New(policy)
//
//	if f.Allow(port, state) {
//		// fire webhook or run command
//	}
package alert
