// Package limiter provides a concurrent-safe sliding-window request limiter
// that tracks per-key call rates and enforces a configurable maximum.
package limiter

import (
	"sync"
	"time"
)

// Policy controls the behaviour of a Limiter.
type Policy struct {
	// Max is the maximum number of calls allowed within Window.
	Max int
	// Window is the duration of the sliding window.
	Window time.Duration
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() Policy {
	return Policy{
		Max:    10,
		Window: time.Minute,
	}
}

type entry struct {
	times []time.Time
}

// Limiter enforces a sliding-window rate limit per string key.
type Limiter struct {
	mu     sync.Mutex
	policy Policy
	keys   map[string]*entry
}

// New creates a Limiter with the supplied Policy.
// Zero values in p are replaced with defaults.
func New(p Policy) *Limiter {
	if p.Max <= 0 {
		p.Max = DefaultPolicy().Max
	}
	if p.Window <= 0 {
		p.Window = DefaultPolicy().Window
	}
	return &Limiter{
		policy: p,
		keys:   make(map[string]*entry),
	}
}

// Allow returns true and records the call if the key has not exceeded its
// limit within the sliding window. It returns false otherwise.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.keys[key]
	if !ok {
		e = &entry{}
		l.keys[key] = e
	}

	cutoff := now.Add(-l.policy.Window)
	valid := e.times[:0]
	for _, t := range e.times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	e.times = valid

	if len(e.times) >= l.policy.Max {
		return false
	}
	e.times = append(e.times, now)
	return true
}

// Remaining returns the number of calls still permitted for key within the
// current window.
func (l *Limiter) Remaining(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.keys[key]
	if !ok {
		return l.policy.Max
	}
	now := time.Now()
	cutoff := now.Add(-l.policy.Window)
	count := 0
	for _, t := range e.times {
		if t.After(cutoff) {
			count++
		}
	}
	r := l.policy.Max - count
	if r < 0 {
		return 0
	}
	return r
}
