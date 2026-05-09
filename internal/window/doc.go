// Package window implements a thread-safe sliding-window counter.
//
// A Window tracks events keyed by an arbitrary string and counts only those
// that occurred within a configurable time window. Older events are pruned
// lazily on every Add or Count call.
//
// Typical usage:
//
//	w := window.New(window.DefaultPolicy())
//	w.Add("port:8080", 1)
//	count := w.Count("port:8080")
//
// The zero-value Policy is not valid; use DefaultPolicy or supply a positive
// Size explicitly.
package window
