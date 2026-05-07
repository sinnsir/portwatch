// Package router dispatches port-state events to registered handlers
// based on a set of configurable routing rules.
package router

import (
	"context"
	"sync"
)

// Event is the minimal interface a routable event must satisfy.
type Event interface {
	Port() int
	State() string
}

// Handler processes a matched event.
type Handler func(ctx context.Context, e Event) error

// Rule pairs a predicate with a handler.
type Rule struct {
	Match   func(Event) bool
	Handler Handler
}

// Router routes events to zero or more handlers whose rules match.
type Router struct {
	mu    sync.RWMutex
	rules []Rule
}

// New returns an empty Router.
func New() *Router {
	return &Router{}
}

// Register appends a routing rule.
func (r *Router) Register(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
}

// Dispatch evaluates every rule against e and calls all matching handlers
// sequentially. The first non-nil error is returned immediately.
func (r *Router) Dispatch(ctx context.Context, e Event) error {
	r.mu.RLock()
	snap := make([]Rule, len(r.rules))
	copy(snap, r.rules)
	r.mu.RUnlock()

	for _, rule := range snap {
		if rule.Match == nil || rule.Match(e) {
			if err := rule.Handler(ctx, e); err != nil {
				return err
			}
		}
	}
	return nil
}

// Len returns the number of registered rules.
func (r *Router) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.rules)
}
