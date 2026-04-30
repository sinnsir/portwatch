package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errBoom = errors.New("boom")

func fastPolicy(attempts int) Policy {
	return Policy{
		MaxAttempts:  attempts,
		InitialDelay: 1 * time.Millisecond,
		Multiplier:   1.0,
		MaxDelay:     10 * time.Millisecond,
	}
}

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := Do(context.Background(), fastPolicy(3), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	calls := 0
	err := Do(context.Background(), fastPolicy(3), func() error {
		calls++
		if calls < 3 {
			return errBoom
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after 3 attempts, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAllAttempts(t *testing.T) {
	calls := 0
	err := Do(context.Background(), fastPolicy(3), func() error {
		calls++
		return errBoom
	})
	if !errors.Is(err, ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected wrapped errBoom, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ContextCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := Do(ctx, fastPolicy(3), func() error {
		calls++
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected 0 calls, got %d", calls)
	}
}

func TestDo_ZeroMaxAttemptsDefaultsToOne(t *testing.T) {
	calls := 0
	p := Policy{MaxAttempts: 0, InitialDelay: time.Millisecond}
	_ = Do(context.Background(), p, func() error {
		calls++
		return errBoom
	})
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}
