// Package router provides a lightweight rule-based event router for portwatch.
//
// Each Rule pairs a predicate (Match) with a Handler. When Dispatch is called
// with an Event, the router evaluates every registered rule in registration
// order and invokes the Handler of every rule whose predicate returns true.
// A nil predicate is treated as an unconditional match.
//
// Dispatch stops and returns the first non-nil error returned by a handler,
// leaving subsequent rules unevaluated.
//
// Router is safe for concurrent use; Register may be called from any goroutine
// while Dispatch is in flight.
package router
