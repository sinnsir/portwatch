// Package retry provides configurable retry logic with exponential backoff
// for use when executing webhooks or shell commands that may transiently fail.
package retry

import (
	"context"
	"errors"
	"time"
)

// Policy defines how retries are performed.
type Policy struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// InitialDelay is the wait time before the second attempt.
	InitialDelay time.Duration
	// Multiplier scales the delay after each failure.
	Multiplier float64
	// MaxDelay caps the computed delay.
	MaxDelay time.Duration
}

// DefaultPolicy returns a sensible retry policy for network operations.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     10 * time.Second,
	}
}

// ErrExhausted is returned when all attempts have been consumed.
var ErrExhausted = errors.New("retry: all attempts exhausted")

// Do executes fn according to p, retrying on non-nil errors.
// The context is checked before every attempt; cancellation stops retries.
func Do(ctx context.Context, p Policy, fn func() error) error {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 1
	}
	delay := p.InitialDelay
	var lastErr error
	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if lastErr = fn(); lastErr == nil {
			return nil
		}
		if attempt == p.MaxAttempts-1 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = time.Duration(float64(delay) * p.Multiplier)
		if p.MaxDelay > 0 && delay > p.MaxDelay {
			delay = p.MaxDelay
		}
	}
	return errors.Join(ErrExhausted, lastErr)
}
