package throttle_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"portwatch/internal/throttle"
)

func TestAllow_ConcurrentSafety(t *testing.T) {
	th := throttle.New(throttle.Policy{
		Rate:           50,
		RefillInterval: time.Hour, // no refill during test
	})

	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if th.Allow() {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()

	if allowed.Load() > 50 {
		t.Fatalf("concurrent calls exceeded rate cap: got %d", allowed.Load())
	}
}

func TestAllow_SteadyStateRefill(t *testing.T) {
	th := throttle.New(throttle.Policy{
		Rate:           2,
		RefillInterval: 30 * time.Millisecond,
	})

	// drain
	th.Allow()
	th.Allow()
	if th.Allow() {
		t.Fatal("bucket should be empty")
	}

	// wait for one refill
	time.Sleep(40 * time.Millisecond)
	if !th.Allow() {
		t.Fatal("expected token after refill interval")
	}
}
