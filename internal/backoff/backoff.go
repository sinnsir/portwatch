// Package backoff provides an exponential back-off calculator used when
// scheduling reconnection attempts or cooldown periods after failures.
package backoff

import (
	"math"
	"time"
)

// Policy controls how delays are computed between successive attempts.
type Policy struct {
	// InitialDelay is the wait time before the second attempt.
	InitialDelay time.Duration
	// Multiplier is applied to the previous delay on each step.
	Multiplier float64
	// MaxDelay caps the computed delay so it never grows unbounded.
	MaxDelay time.Duration
	// Jitter adds a random fraction (0–1) of the computed delay to spread load.
	Jitter bool
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() Policy {
	return Policy{
		InitialDelay: 500 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     30 * time.Second,
		Jitter:       true,
	}
}

// Calculator holds state for a single back-off sequence.
type Calculator struct {
	policy  Policy
	attempt int
	rng     func() float64 // returns value in [0, 1); injectable for tests
}

// New returns a Calculator using p. If p.Multiplier <= 1 it is set to 2.
// If p.InitialDelay <= 0 it is set to 100 ms.
func New(p Policy) *Calculator {
	if p.Multiplier <= 1 {
		p.Multiplier = 2
	}
	if p.InitialDelay <= 0 {
		p.InitialDelay = 100 * time.Millisecond
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = 30 * time.Second
	}
	return &Calculator{policy: p, rng: defaultRng}
}

// Next returns the delay for the current attempt and advances the internal
// counter. The first call (attempt 0) always returns 0 so the initial
// execution is immediate.
func (c *Calculator) Next() time.Duration {
	if c.attempt == 0 {
		c.attempt++
		return 0
	}
	exp := math.Pow(c.policy.Multiplier, float64(c.attempt-1))
	d := time.Duration(float64(c.policy.InitialDelay) * exp)
	if d > c.policy.MaxDelay {
		d = c.policy.MaxDelay
	}
	if c.policy.Jitter {
		jitter := time.Duration(float64(d) * c.rng())
		d += jitter
		if d > c.policy.MaxDelay {
			d = c.policy.MaxDelay
		}
	}
	c.attempt++
	return d
}

// Reset restarts the sequence from attempt 0.
func (c *Calculator) Reset() { c.attempt = 0 }

// Attempt returns the current attempt index (0-based).
func (c *Calculator) Attempt() int { return c.attempt }
