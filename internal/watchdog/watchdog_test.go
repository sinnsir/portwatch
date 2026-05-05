package watchdog_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/watchdog"
)

func fastPolicy() watchdog.Policy {
	return watchdog.Policy{
		Interval:  20 * time.Millisecond,
		Threshold: 2,
	}
}

func TestWatchdog_NoRestartWhenAlive(t *testing.T) {
	var restarts int64
	wd := watchdog.New(fastPolicy(), func() bool { return true }, func() {
		atomic.AddInt64(&restarts, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	wd.Run(ctx)

	if r := atomic.LoadInt64(&restarts); r != 0 {
		t.Fatalf("expected 0 restarts, got %d", r)
	}
}

func TestWatchdog_RestartAfterThreshold(t *testing.T) {
	var restarts int64
	wd := watchdog.New(fastPolicy(), func() bool { return false }, func() {
		atomic.AddInt64(&restarts, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	wd.Run(ctx)

	if r := atomic.LoadInt64(&restarts); r < 1 {
		t.Fatalf("expected at least 1 restart, got %d", r)
	}
}

func TestWatchdog_FailureCountResetsAfterRecovery(t *testing.T) {
	calls := int64(0)
	probe := func() bool {
		n := atomic.AddInt64(&calls, 1)
		return n%2 == 0 // alternates: fail, pass, fail, pass …
	}

	var restarts int64
	wd := watchdog.New(fastPolicy(), probe, func() {
		atomic.AddInt64(&restarts, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	wd.Run(ctx)

	// With alternating probe, threshold of 2 should never be reached.
	if r := atomic.LoadInt64(&restarts); r != 0 {
		t.Fatalf("expected 0 restarts with alternating probe, got %d", r)
	}
}

func TestWatchdog_DefaultPolicyUsedOnZeroValues(t *testing.T) {
	p := watchdog.Policy{} // zero values
	wd := watchdog.New(p, func() bool { return true }, func() {})
	if wd == nil {
		t.Fatal("expected non-nil watchdog")
	}
}

func TestWatchdog_FailuresAccessible(t *testing.T) {
	probe := func() bool { return false }
	wd := watchdog.New(fastPolicy(), probe, func() {})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	wd.Run(ctx)

	// After running, failures should have been recorded (then reset on restart).
	// We just assert the method is callable without panic.
	_ = wd.Failures()
}
