// Package sampler provides probabilistic event sampling for portwatch.
// It allows a configurable fraction of events to pass through, reducing
// downstream load while preserving statistical representativeness.
package sampler

import (
	"math/rand"
	"sync"
	"time"
)

// Policy controls sampler behaviour.
type Policy struct {
	// Rate is the fraction of events to allow through, in the range (0, 1].
	// A value of 1.0 passes every event; 0.1 passes roughly one in ten.
	Rate float64
}

// DefaultPolicy returns a Policy that passes every event.
func DefaultPolicy() Policy {
	return Policy{Rate: 1.0}
}

// Sampler decides probabilistically whether an event should be processed.
type Sampler struct {
	mu   sync.Mutex
	rng  *rand.Rand
	rate float64
}

// New returns a Sampler configured with the given Policy.
// If p.Rate is outside (0, 1] it is clamped to 1.0.
func New(p Policy) *Sampler {
	rate := p.Rate
	if rate <= 0 || rate > 1 {
		rate = 1.0
	}
	return &Sampler{
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		rate: rate,
	}
}

// Allow returns true if the event should be processed according to the
// configured sample rate. It is safe for concurrent use.
func (s *Sampler) Allow() bool {
	if s.rate >= 1.0 {
		return true
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.rate
}

// Rate returns the configured sample rate.
func (s *Sampler) Rate() float64 {
	return s.rate
}
