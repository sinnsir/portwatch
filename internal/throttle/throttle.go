// Package throttle provides a token-bucket style throughput limiter that
// caps the number of actions dispatched per unit of time across all monitored
// ports. Unlike ratelimit (per-key fixed-window) this bucket refills
// continuously and is shared globally, making it suitable for protecting
// downstream webhook receivers from burst traffic.
package throttle

import (
	"sync"
	"time"
)

// Policy controls the throttle behaviour.
type Policy struct {
	// Rate is the maximum number of tokens available in the bucket.
	Rate int
	// RefillInterval is how often one token is added back to the bucket.
	RefillInterval time.Duration
}

// DefaultPolicy returns a sensible out-of-the-box policy: 10 tokens, one
// refilled every 500 ms (≈ 20 actions/s burst, 2 actions/s steady-state).
func DefaultPolicy() Policy {
	return Policy{
		Rate:           10,
		RefillInterval: 500 * time.Millisecond,
	}
}

// Throttle is a thread-safe token-bucket limiter.
type Throttle struct {
	mu     sync.Mutex
	tokens int
	policy Policy
	last   time.Time
	now    func() time.Time // injectable for tests
}

// New constructs a Throttle with the given policy. A zero Policy falls back to
// DefaultPolicy.
func New(p Policy) *Throttle {
	if p.Rate <= 0 {
		p.Rate = DefaultPolicy().Rate
	}
	if p.RefillInterval <= 0 {
		p.RefillInterval = DefaultPolicy().RefillInterval
	}
	return &Throttle{
		tokens: p.Rate,
		policy: p,
		last:   time.Now(),
		now:    time.Now,
	}
}

// Allow returns true and consumes one token when the bucket is non-empty.
// It refills tokens proportionally to elapsed time before deciding.
func (t *Throttle) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	elapsed := now.Sub(t.last)
	refill := int(elapsed / t.policy.RefillInterval)
	if refill > 0 {
		t.tokens += refill
		if t.tokens > t.policy.Rate {
			t.tokens = t.policy.Rate
		}
		t.last = t.last.Add(time.Duration(refill) * t.policy.RefillInterval)
	}

	if t.tokens <= 0 {
		return false
	}
	t.tokens--
	return true
}

// Remaining returns the current number of available tokens without consuming
// one. Useful for metrics and debugging.
func (t *Throttle) Remaining() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tokens
}
