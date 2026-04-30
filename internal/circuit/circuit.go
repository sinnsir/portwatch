// Package circuit implements a simple circuit breaker for outbound actions
// (webhooks, shell commands). It prevents repeated calls to a failing target
// by opening after a configurable failure threshold and self-healing after a
// cooldown period.
package circuit

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking calls
	StateHalfOpen              // probing for recovery
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrOpen is returned when a call is rejected because the circuit is open.
var ErrOpen = errors.New("circuit breaker is open")

// Policy configures the circuit breaker behaviour.
type Policy struct {
	MaxFailures int
	Cooldown    time.Duration
}

// DefaultPolicy returns sensible defaults.
func DefaultPolicy() Policy {
	return Policy{
		MaxFailures: 5,
		Cooldown:    30 * time.Second,
	}
}

// Breaker is a per-key circuit breaker.
type Breaker struct {
	policy  Policy
	mu      sync.Mutex
	state   State
	failures int
	openedAt time.Time
}

// New creates a Breaker with the given policy.
func New(p Policy) *Breaker {
	if p.MaxFailures <= 0 {
		p.MaxFailures = 1
	}
	if p.Cooldown <= 0 {
		p.Cooldown = time.Second
	}
	return &Breaker{policy: p}
}

// Allow returns nil if the call should proceed, or ErrOpen if it is blocked.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case StateOpen:
		if time.Since(b.openedAt) >= b.policy.Cooldown {
			b.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	default:
		return nil
	}
}

// RecordSuccess resets failure count and closes the circuit.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and may open the circuit.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.policy.MaxFailures {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// State returns the current circuit state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
