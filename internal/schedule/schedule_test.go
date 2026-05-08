package schedule

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func fastPolicy(interval time.Duration) Policy {
	return Policy{Interval: interval, SkipIfBusy: true}
}

func TestRun_JobCalledOnTick(t *testing.T) {
	var count int64
	sched := New(fastPolicy(20*time.Millisecond), func(_ context.Context) {
		atomic.AddInt64(&count, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Millisecond)
	defer cancel()
	sched.Run(ctx)

	got := atomic.LoadInt64(&count)
	if got < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", got)
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	var count int64
	sched := New(fastPolicy(10*time.Millisecond), func(_ context.Context) {
		atomic.AddInt64(&count, 1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		sched.Run(ctx)
		close(done)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not return after context cancel")
	}
}

func TestRun_SkipIfBusy_NoOverlap(t *testing.T) {
	var concurrent int64
	var overlap int64

	sched := New(fastPolicy(10*time.Millisecond), func(_ context.Context) {
		v := atomic.AddInt64(&concurrent, 1)
		if v > 1 {
			atomic.AddInt64(&overlap, 1)
		}
		time.Sleep(25 * time.Millisecond) // longer than interval
		atomic.AddInt64(&concurrent, -1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	sched.Run(ctx)

	if atomic.LoadInt64(&overlap) > 0 {
		t.Fatal("overlapping executions detected with SkipIfBusy=true")
	}
}

func TestDefaultPolicy_HasPositiveInterval(t *testing.T) {
	p := DefaultPolicy()
	if p.Interval <= 0 {
		t.Fatalf("expected positive interval, got %v", p.Interval)
	}
}

func TestNew_ZeroIntervalUsesDefault(t *testing.T) {
	sched := New(Policy{Interval: 0}, func(_ context.Context) {})
	if sched.policy.Interval != DefaultPolicy().Interval {
		t.Fatalf("expected default interval, got %v", sched.policy.Interval)
	}
}
