// Package tagger assigns tags to port events based on configurable rules.
// Tags can later be consumed by the filter pipeline to route or suppress events.
package tagger

import (
	"strings"
	"sync"

	"github.com/user/portwatch/internal/portcheck"
)

// Rule maps a predicate to a set of tags that should be applied when it matches.
type Rule struct {
	// Ports is an optional list of port numbers the rule applies to.
	// An empty slice means the rule applies to all ports.
	Ports []int
	// States is an optional list of states the rule applies to.
	// An empty slice means the rule applies to all states.
	States []portcheck.State
	// Tags are the labels attached to matching events.
	Tags []string
}

// Event is the minimal interface the tagger needs from an incoming event.
type Event struct {
	Port  int
	State portcheck.State
	Tags  []string
}

// Tagger applies tag rules to events.
type Tagger struct {
	mu    sync.RWMutex
	rules []Rule
}

// New returns a Tagger pre-loaded with the provided rules.
func New(rules []Rule) *Tagger {
	return &Tagger{rules: rules}
}

// AddRule appends a rule at runtime. It is safe for concurrent use.
func (t *Tagger) AddRule(r Rule) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rules = append(t.rules, r)
}

// Apply returns a copy of the event with all matching tags merged in.
// Duplicate tags are deduplicated; original tags are preserved.
func (t *Tagger) Apply(e Event) Event {
	t.mu.RLock()
	defer t.mu.RUnlock()

	seen := make(map[string]struct{}, len(e.Tags))
	out := make([]string, len(e.Tags))
	copy(out, e.Tags)
	for _, tag := range e.Tags {
		seen[tag] = struct{}{}
	}

	for _, rule := range t.rules {
		if !t.matchesPort(rule, e.Port) || !t.matchesState(rule, e.State) {
			continue
		}
		for _, tag := range rule.Tags {
			norm := strings.TrimSpace(tag)
			if norm == "" {
				continue
			}
			if _, exists := seen[norm]; !exists {
				seen[norm] = struct{}{}
				out = append(out, norm)
			}
		}
	}

	e.Tags = out
	return e
}

func (t *Tagger) matchesPort(r Rule, port int) bool {
	if len(r.Ports) == 0 {
		return true
	}
	for _, p := range r.Ports {
		if p == port {
			return true
		}
	}
	return false
}

func (t *Tagger) matchesState(r Rule, state portcheck.State) bool {
	if len(r.States) == 0 {
		return true
	}
	for _, s := range r.States {
		if s == state {
			return true
		}
	}
	return false
}
