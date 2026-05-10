package cooldown_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/cooldown"
)

func TestAllow_ConcurrentSafety(t *testing.T) {
	tr := cooldown.New(cooldown.Policy{Window: 10 * time.Millisecond})
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			tr.Allow("shared-key")
		}()
	}
	wg.Wait() // must not panic or deadlock
}

func TestAllow_CooldownExpiry_AllowsExactlyOnePerWindow(t *testing.T) {
	window := 40 * time.Millisecond
	tr := cooldown.New(cooldown.Policy{Window: window})

	var allowed atomic.Int64
	const cycles = 3
	for i := 0; i < cycles; i++ {
		if tr.Allow("k") {
			allowed.Add(1)
		}
		// second call in same window must be suppressed
		tr.Allow("k")
		time.Sleep(window + 10*time.Millisecond)
	}
	if allowed.Load() != cycles {
		t.Fatalf("expected %d allowed calls, got %d", cycles, allowed.Load())
	}
}

func TestAllow_ManyKeys_IndependentCooldowns(t *testing.T) {
	tr := cooldown.New(cooldown.Policy{Window: 100 * time.Millisecond})
	const n = 20
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("port:%d", 8000+i)
		if !tr.Allow(key) {
			t.Fatalf("key %s: expected first call to pass", key)
		}
	}
	// All keys are now in cooldown; none should pass.
	for i := 0; i < n; i++ {
		key := fmt.Sprintf("port:%d", 8000+i)
		if tr.Allow(key) {
			t.Fatalf("key %s: expected suppression within window", key)
		}
	}
}
