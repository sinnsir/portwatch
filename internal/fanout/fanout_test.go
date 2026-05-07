package fanout_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/fanout"
)

func TestPublish_NoHandlers_ReturnsNil(t *testing.T) {
	f := fanout.New[string]()
	errs := f.Publish(context.Background(), "event")
	if errs != nil {
		t.Fatalf("expected nil errors, got %v", errs)
	}
}

func TestPublish_AllHandlersCalled(t *testing.T) {
	f := fanout.New[int]()
	var count atomic.Int32

	for i := 0; i < 5; i++ {
		f.Register(func(_ context.Context, _ int) error {
			count.Add(1)
			return nil
		})
	}

	errs := f.Publish(context.Background(), 42)
	if errs != nil {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if got := count.Load(); got != 5 {
		t.Fatalf("expected 5 calls, got %d", got)
	}
}

func TestPublish_ErrorsCollected(t *testing.T) {
	f := fanout.New[string]()
	errA := errors.New("handler A failed")
	errB := errors.New("handler B failed")

	f.Register(func(_ context.Context, _ string) error { return errA })
	f.Register(func(_ context.Context, _ string) error { return nil })
	f.Register(func(_ context.Context, _ string) error { return errB })

	errs := f.Publish(context.Background(), "x")
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestPublish_CancelledContextReturnsImmediately(t *testing.T) {
	f := fanout.New[string]()
	var called atomic.Bool
	f.Register(func(_ context.Context, _ string) error {
		called.Store(true)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	errs := f.Publish(ctx, "event")
	if len(errs) == 0 {
		t.Fatal("expected context error, got none")
	}
	if called.Load() {
		t.Fatal("handler should not have been called")
	}
}

func TestPublish_ConcurrentHandlers(t *testing.T) {
	f := fanout.New[int]()
	const n = 20
	var count atomic.Int32

	for i := 0; i < n; i++ {
		f.Register(func(_ context.Context, _ int) error {
			time.Sleep(2 * time.Millisecond)
			count.Add(1)
			return nil
		})
	}

	start := time.Now()
	f.Publish(context.Background(), 1)
	elapsed := time.Since(start)

	if count.Load() != n {
		t.Fatalf("expected %d calls, got %d", n, count.Load())
	}
	// All handlers run concurrently; total time should be well under n*2ms.
	if elapsed > 100*time.Millisecond {
		t.Fatalf("handlers did not run concurrently: elapsed %v", elapsed)
	}
}

func TestLen_ReflectsRegistrations(t *testing.T) {
	f := fanout.New[bool]()
	if f.Len() != 0 {
		t.Fatalf("expected 0, got %d", f.Len())
	}
	f.Register(func(_ context.Context, _ bool) error { return nil })
	f.Register(func(_ context.Context, _ bool) error { return nil })
	if f.Len() != 2 {
		t.Fatalf("expected 2, got %d", f.Len())
	}
}
