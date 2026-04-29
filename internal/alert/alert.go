// Package alert provides rate-limiting and deduplication for port state change notifications.
package alert

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/portcheck"
)

// Policy controls when alerts are suppressed.
type Policy struct {
	// Cooldown is the minimum duration between repeated alerts for the same port+state.
	Cooldown time.Duration
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() Policy {
	return Policy{Cooldown: 30 * time.Second}
}

type alertKey struct {
	port  int
	state portcheck.State
}

// Filter tracks recent alerts and suppresses duplicates within the cooldown window.
type Filter struct {
	mu     sync.Mutex
	policy Policy
	last   map[alertKey]time.Time
	now    func() time.Time
}

// New creates a Filter with the given Policy.
func New(p Policy) *Filter {
	return &Filter{
		policy: p,
		last:   make(map[alertKey]time.Time),
		now:    time.Now,
	}
}

// Allow returns true if the alert for the given port and state should be fired.
// It records the current time so subsequent calls within the cooldown return false.
func (f *Filter) Allow(port int, state portcheck.State) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := alertKey{port: port, state: state}
	now := f.now()

	if last, ok := f.last[key]; ok {
		if now.Sub(last) < f.policy.Cooldown {
			return false
		}
	}

	f.last[key] = now
	return true
}

// Reset clears the recorded alert time for a specific port and state.
func (f *Filter) Reset(port int, state portcheck.State) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.last, alertKey{port: port, state: state})
}
