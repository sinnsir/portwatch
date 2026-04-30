package ratelimit_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"portwatch/internal/ratelimit"
)

func TestAllow_ConcurrentSafety(t *testing.T) {
	l := ratelimit.New(50, time.Minute)
	key := "concurrent"

	var allowed atomic.Int64
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if l.Allow(key) {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if got := allowed.Load(); got != 50 {
		t.Fatalf("expected exactly 50 allowed, got %d", got)
	}
}

func TestAllow_MultipleWindowCycles(t *testing.T) {
	l := ratelimit.New(2, 40*time.Millisecond)
	key := "cycles"

	for cycle := 0; cycle < 3; cycle++ {
		passed := 0
		for i := 0; i < 5; i++ {
			if l.Allow(key) {
				passed++
			}
		}
		if passed != 2 {
			t.Fatalf("cycle %d: expected 2 allowed, got %d", cycle, passed)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
