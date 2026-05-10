// Package cooldown provides a per-key cooldown tracker that suppresses
// repeated signals within a configurable quiet period.
package cooldown

import (
	"sync"
	"time"
)

// Policy configures cooldown behaviour.
type Policy struct {
	// Window is the minimum duration that must elapse between two allowed
	// signals for the same key.
	Window time.Duration
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() Policy {
	return Policy{Window: 30 * time.Second}
}

func (p Policy) withDefaults() Policy {
	if p.Window <= 0 {
		p.Window = DefaultPolicy().Window
	}
	return p
}

// Tracker suppresses duplicate signals within a cooldown window.
type Tracker struct {
	policy Policy
	mu     sync.Mutex
	last   map[string]time.Time
	now    func() time.Time
}

// New returns a Tracker using the given Policy.
func New(p Policy) *Tracker {
	return &Tracker{
		policy: p.withDefaults(),
		last:   make(map[string]time.Time),
		now:    time.Now,
	}
}

// Allow returns true if the cooldown window for key has elapsed since the
// last allowed call. It always returns true on the first call for a key.
func (t *Tracker) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if last, ok := t.last[key]; ok {
		if now.Sub(last) < t.policy.Window {
			return false
		}
	}
	t.last[key] = now
	return true
}

// Reset clears the cooldown state for key, allowing the next call to pass
// immediately.
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.last, key)
}

// Remaining returns how long until the cooldown for key expires. It returns
// zero if the key is not in cooldown.
func (t *Tracker) Remaining(key string) time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()

	last, ok := t.last[key]
	if !ok {
		return 0
	}
	remaining := t.policy.Window - t.now().Sub(last)
	if remaining < 0 {
		return 0
	}
	return remaining
}
