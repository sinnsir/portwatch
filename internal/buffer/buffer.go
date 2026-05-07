// Package buffer provides a bounded, thread-safe event buffer that accumulates
// events and flushes them in batches either when the buffer is full or a flush
// interval elapses.
package buffer

import (
	"context"
	"sync"
	"time"
)

// Policy controls buffer behaviour.
type Policy struct {
	// Capacity is the maximum number of items held before an automatic flush.
	Capacity int
	// FlushInterval is the maximum time between flushes.
	FlushInterval time.Duration
}

// DefaultPolicy returns a sensible out-of-the-box policy.
func DefaultPolicy() Policy {
	return Policy{
		Capacity:      64,
		FlushInterval: 5 * time.Second,
	}
}

// FlushFunc is called with a snapshot of buffered items on every flush.
type FlushFunc[T any] func(items []T)

// Buffer accumulates items and flushes them in batches.
type Buffer[T any] struct {
	mu       sync.Mutex
	policy   Policy
	items    []T
	flushFn  FlushFunc[T]
	ticker   *time.Ticker
	stopOnce sync.Once
	stopCh   chan struct{}
}

// New creates a Buffer with the given policy and flush callback.
// Call Run to start the background timer goroutine.
func New[T any](p Policy, fn FlushFunc[T]) *Buffer[T] {
	if p.Capacity <= 0 {
		p.Capacity = DefaultPolicy().Capacity
	}
	if p.FlushInterval <= 0 {
		p.FlushInterval = DefaultPolicy().FlushInterval
	}
	return &Buffer[T]{
		policy:  p,
		items:   make([]T, 0, p.Capacity),
		flushFn: fn,
		stopCh:  make(chan struct{}),
	}
}

// Add appends an item to the buffer. If the buffer reaches capacity it is
// flushed immediately.
func (b *Buffer[T]) Add(item T) {
	b.mu.Lock()
	b.items = append(b.items, item)
	should := len(b.items) >= b.policy.Capacity
	b.mu.Unlock()
	if should {
		b.Flush()
	}
}

// Flush drains the buffer and calls the flush function synchronously.
func (b *Buffer[T]) Flush() {
	b.mu.Lock()
	if len(b.items) == 0 {
		b.mu.Unlock()
		return
	}
	snapshot := make([]T, len(b.items))
	copy(snapshot, b.items)
	b.items = b.items[:0]
	b.mu.Unlock()
	b.flushFn(snapshot)
}

// Run starts the background ticker that triggers periodic flushes.
// It blocks until ctx is cancelled or Stop is called.
func (b *Buffer[T]) Run(ctx context.Context) {
	b.ticker = time.NewTicker(b.policy.FlushInterval)
	defer b.ticker.Stop()
	for {
		select {
		case <-b.ticker.C:
			b.Flush()
		case <-ctx.Done():
			b.Flush()
			return
		case <-b.stopCh:
			b.Flush()
			return
		}
	}
}

// Stop signals the background goroutine to exit and performs a final flush.
func (b *Buffer[T]) Stop() {
	b.stopOnce.Do(func() { close(b.stopCh) })
}
