// Package backoff provides jittered exponential backoff calculation
// for use in retry and circuit-breaker logic.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Policy controls how backoff durations are computed.
type Policy struct {
	// InitialDelay is the base delay for the first backoff step.
	InitialDelay time.Duration
	// Multiplier is applied to the delay on each successive step.
	Multiplier float64
	// MaxDelay caps the computed delay regardless of step count.
	MaxDelay time.Duration
	// Jitter adds random noise as a fraction of the computed delay (0–1).
	Jitter float64
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() Policy {
	return Policy{
		InitialDelay: 500 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     30 * time.Second,
		Jitter:       0.2,
	}
}

// Duration returns the backoff duration for the given attempt (0-indexed).
// It applies exponential growth, caps at MaxDelay, then adds jitter.
func (p Policy) Duration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	base := float64(p.InitialDelay) * math.Pow(p.Multiplier, float64(attempt))
	if base > float64(p.MaxDelay) {
		base = float64(p.MaxDelay)
	}

	if p.Jitter > 0 {
		// jitter is ± (Jitter/2) * base
		noise := (rand.Float64() - 0.5) * p.Jitter * base
		base += noise
		if base < 0 {
			base = 0
		}
	}

	return time.Duration(base)
}
