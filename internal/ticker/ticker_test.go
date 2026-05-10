package ticker

import (
	"context"
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{
		Interval: 20 * time.Millisecond,
		Jitter:   0,
	}
}

func TestNew_DefaultIntervalOnZero(t *testing.T) {
	tk := New(Policy{})
	if tk.policy.Interval != DefaultPolicy().Interval {
		t.Fatalf("expected default interval %v, got %v", DefaultPolicy().Interval, tk.policy.Interval)
	}
}

func TestNew_NegativeJitterClampsToZero(t *testing.T) {
	tk := New(Policy{Interval: time.Second, Jitter: -1})
	if tk.policy.Jitter != 0 {
		t.Fatalf("expected jitter 0, got %v", tk.policy.Jitter)
	}
}

func TestDefaultPolicy_HasPositiveInterval(t *testing.T) {
	p := DefaultPolicy()
	if p.Interval <= 0 {
		t.Fatal("default interval must be positive")
	}
}

func TestDefaultPolicy_HasNonNegativeJitter(t *testing.T) {
	p := DefaultPolicy()
	if p.Jitter < 0 {
		t.Fatal("default jitter must be non-negative")
	}
}

func TestRun_EmitsTicks(t *testing.T) {
	tk := New(fastPolicy())
	ch := make(chan time.Time, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go tk.Run(ctx, ch)

	count := 0
	for {
		select {
		case <-ch:
			count++
		case <-ctx.Done():
			if count < 3 {
				t.Fatalf("expected at least 3 ticks, got %d", count)
			}
			return
		}
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	tk := New(fastPolicy())
	ch := make(chan time.Time, 2)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		tk.Run(ctx, ch)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not stop after context cancel")
	}
}

func TestRun_WithJitter_StillEmits(t *testing.T) {
	tk := New(Policy{Interval: 20 * time.Millisecond, Jitter: 5 * time.Millisecond})
	ch := make(chan time.Time, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	go tk.Run(ctx, ch)

	select {
	case <-ch:
		// at least one tick received
	case <-ctx.Done():
		t.Fatal("no ticks received with jitter enabled")
	}
}
