// Package ticker provides a reusable interval ticker that emits ticks on a
// channel and can be stopped cleanly. It wraps time.Ticker with context
// awareness and configurable jitter to spread load across multiple instances.
package ticker

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Policy controls ticker behaviour.
type Policy struct {
	// Interval is the base duration between ticks.
	Interval time.Duration
	// Jitter is the maximum random duration added to each interval.
	// Set to zero to disable jitter.
	Jitter time.Duration
}

// DefaultPolicy returns a sensible default policy.
func DefaultPolicy() Policy {
	return Policy{
		Interval: 10 * time.Second,
		Jitter:   500 * time.Millisecond,
	}
}

// Ticker emits ticks at a configurable interval.
type Ticker struct {
	policy Policy
	mu     sync.Mutex
	rng    *rand.Rand
}

// New creates a new Ticker with the given policy.
// Zero-value interval fields fall back to DefaultPolicy values.
func New(p Policy) *Ticker {
	def := DefaultPolicy()
	if p.Interval <= 0 {
		p.Interval = def.Interval
	}
	if p.Jitter < 0 {
		p.Jitter = 0
	}
	return &Ticker{
		policy: p,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Run blocks, sending the current time on ch at each interval until ctx is
// cancelled. The channel is not closed on exit; callers should select on
// ctx.Done() alongside the channel.
func (t *Ticker) Run(ctx context.Context, ch chan<- time.Time) {
	for {
		delay := t.next()
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		select {
		case ch <- time.Now():
		case <-ctx.Done():
			return
		}
	}
}

// next returns the duration until the next tick, including any jitter.
func (t *Ticker) next() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.policy.Jitter == 0 {
		return t.policy.Interval
	}
	jitter := time.Duration(t.rng.Int63n(int64(t.policy.Jitter)))
	return t.policy.Interval + jitter
}
