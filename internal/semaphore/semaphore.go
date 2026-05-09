// Package semaphore provides a counting semaphore for bounding concurrent
// goroutines. It is safe for concurrent use.
package semaphore

import (
	"context"
	"errors"
)

// ErrAcquire is returned when Acquire fails due to a cancelled context.
var ErrAcquire = errors.New("semaphore: acquire cancelled")

// Semaphore limits the number of concurrent operations.
type Semaphore struct {
	slots chan struct{}
}

// New returns a Semaphore that allows at most n concurrent acquisitions.
// If n < 1 it is clamped to 1.
func New(n int) *Semaphore {
	if n < 1 {
		n = 1
	}
	s := &Semaphore{slots: make(chan struct{}, n)}
	for i := 0; i < n; i++ {
		s.slots <- struct{}{}
	}
	return s
}

// Acquire blocks until a slot is available or ctx is done.
// Returns ErrAcquire if the context is cancelled before a slot is obtained.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-s.slots:
		return nil
	case <-ctx.Done():
		return ErrAcquire
	}
}

// TryAcquire attempts to acquire a slot without blocking.
// Returns true on success, false if no slot is immediately available.
func (s *Semaphore) TryAcquire() bool {
	select {
	case <-s.slots:
		return true
	default:
		return false
	}
}

// Release returns a previously acquired slot back to the semaphore.
func (s *Semaphore) Release() {
	s.slots <- struct{}{}
}

// Available returns the number of slots currently available.
func (s *Semaphore) Available() int {
	return len(s.slots)
}

// Cap returns the total capacity of the semaphore.
func (s *Semaphore) Cap() int {
	return cap(s.slots)
}
