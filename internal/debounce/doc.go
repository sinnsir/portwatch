// Package debounce suppresses transient port-state flaps before they reach
// the notifier and action pipeline.
//
// A Debouncer tracks per-key (host:port) consecutive observations.  Only
// when the same boolean state (open/closed) is seen at least Threshold times
// within Window does Observe return true, signalling that the state change is
// stable enough to act upon.
//
// Typical usage inside a monitor loop:
//
//	d := debounce.New(debounce.DefaultPolicy())
//	...
//	if d.Observe(key, isOpen) {
//		// forward stable state change to notifier
//	}
package debounce
