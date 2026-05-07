// Package fanout broadcasts a single event to multiple handlers concurrently.
// Each handler runs in its own goroutine; errors are collected and returned
// as a combined slice after all handlers finish.
package fanout

import (
	"context"
	"sync"
)

// Handler is a function that processes an event of any type.
type Handler[T any] func(ctx context.Context, event T) error

// Fanout broadcasts events to a registered set of handlers concurrently.
type Fanout[T any] struct {
	mu       sync.RWMutex
	handlers []Handler[T]
}

// New returns an empty Fanout ready for use.
func New[T any]() *Fanout[T] {
	return &Fanout[T]{}
}

// Register adds a handler to the broadcast list.
// Safe to call after construction; concurrent calls are serialised.
func (f *Fanout[T]) Register(h Handler[T]) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers = append(f.handlers, h)
}

// Len returns the number of registered handlers.
func (f *Fanout[T]) Len() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.handlers)
}

// Publish sends event to every registered handler concurrently.
// It waits for all handlers to finish and returns all non-nil errors.
// If ctx is already cancelled before Publish is called, it returns
// immediately with ctx.Err().
func (f *Fanout[T]) Publish(ctx context.Context, event T) []error {
	if err := ctx.Err(); err != nil {
		return []error{err}
	}

	f.mu.RLock()
	snap := make([]Handler[T], len(f.handlers))
	copy(snap, f.handlers)
	f.mu.RUnlock()

	if len(snap) == 0 {
		return nil
	}

	type result struct {
		err error
	}

	results := make([]result, len(snap))
	var wg sync.WaitGroup
	wg.Add(len(snap))

	for i, h := range snap {
		i, h := i, h
		go func() {
			defer wg.Done()
			results[i].err = h(ctx, event)
		}()
	}

	wg.Wait()

	var errs []error
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}
	return errs
}
