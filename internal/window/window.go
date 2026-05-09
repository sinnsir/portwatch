// Package window provides a sliding-window counter keyed by an arbitrary
// string. It is useful for rate-limiting, anomaly detection, or any scenario
// where you need to count events that occurred within the last N seconds.
package window

import (
	"sync"
	"time"
)

// Policy controls the behaviour of a Window.
type Policy struct {
	// Size is the duration of the sliding window.
	Size time.Duration
}

// DefaultPolicy returns a Policy with sensible defaults.
func DefaultPolicy() Policy {
	return Policy{Size: 60 * time.Second}
}

type bucket struct {
	at    time.Time
	count int64
}

// Window is a thread-safe sliding-window counter.
type Window struct {
	mu      sync.Mutex
	policy  Policy
	buckets map[string][]bucket
	now     func() time.Time // injectable for testing
}

// New creates a new Window with the given policy.
// If policy.Size is <= 0 it is reset to the default.
func New(p Policy) *Window {
	if p.Size <= 0 {
		p.Size = DefaultPolicy().Size
	}
	return &Window{
		policy:  p,
		buckets: make(map[string][]bucket),
		now:     time.Now,
	}
}

// Add records delta events for key at the current time and returns the total
// count within the window after the addition.
func (w *Window) Add(key string, delta int64) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	w.buckets[key] = append(w.prune(key, now), bucket{at: now, count: delta})
	return w.sum(key)
}

// Count returns the total number of events recorded for key within the window.
func (w *Window) Count(key string) int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	w.buckets[key] = w.prune(key, now)
	return w.sum(key)
}

// Reset clears all recorded events for key.
func (w *Window) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.buckets, key)
}

// prune removes buckets outside the window; caller must hold mu.
func (w *Window) prune(key string, now time.Time) []bucket {
	cutoff := now.Add(-w.policy.Size)
	bs := w.buckets[key]
	i := 0
	for i < len(bs) && bs[i].at.Before(cutoff) {
		i++
	}
	return bs[i:]
}

// sum totals all bucket counts for key; caller must hold mu.
func (w *Window) sum(key string) int64 {
	var total int64
	for _, b := range w.buckets[key] {
		total += b.count
	}
	return total
}
