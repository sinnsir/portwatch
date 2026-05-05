package retry_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/retry"
)

func TestDo_ContextCancelledDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	p := retry.Policy{
		MaxAttempts:  10,
		InitialDelay: 50 * time.Millisecond, // longer than ctx timeout
		Multiplier:   1.0,
		MaxDelay:     1 * time.Second,
	}

	calls := 0
	err := retry.Do(ctx, p, func() error {
		calls++
		return errors.New("transient")
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 call before backoff cancelled, got %d", calls)
	}
}

func TestDo_ConcurrentSafety(t *testing.T) {
	p := retry.Policy{
		MaxAttempts:  2,
		InitialDelay: 1 * time.Millisecond,
		Multiplier:   1.0,
		MaxDelay:     5 * time.Millisecond,
	}

	var total atomic.Int64
	const goroutines = 20
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			errs <- retry.Do(context.Background(), p, func() error {
				total.Add(1)
				return nil
			})
		}()
	}

	for i := 0; i < goroutines; i++ {
		if err := <-errs; err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}
	if total.Load() != goroutines {
		t.Fatalf("expected %d calls, got %d", goroutines, total.Load())
	}
}

func TestDo_ExhaustsMaxAttempts(t *testing.T) {
	p := retry.Policy{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		Multiplier:   1.0,
		MaxDelay:     5 * time.Millisecond,
	}

	sentinel := errors.New("permanent failure")
	calls := 0
	err := retry.Do(context.Background(), p, func() error {
		calls++
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != p.MaxAttempts {
		t.Fatalf("expected %d calls, got %d", p.MaxAttempts, calls)
	}
}
