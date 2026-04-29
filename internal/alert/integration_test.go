package alert_test

import (
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/portcheck"
)

// TestAllow_ConcurrentSafety verifies the filter does not race under concurrent access.
func TestAllow_ConcurrentSafety(t *testing.T) {
	f := alert.New(alert.Policy{Cooldown: 5 * time.Millisecond})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			f.Allow(port, portcheck.StateOpen)
			f.Allow(port, portcheck.StateClosed)
			f.Reset(port, portcheck.StateOpen)
		}(i % 5)
	}
	wg.Wait()
}

// TestAllow_CooldownExpiry confirms that an alert is re-allowed after the cooldown elapses.
func TestAllow_CooldownExpiry(t *testing.T) {
	cooldown := 20 * time.Millisecond
	f := alert.New(alert.Policy{Cooldown: cooldown})

	if !f.Allow(3000, portcheck.StateOpen) {
		t.Fatal("first allow should pass")
	}
	if f.Allow(3000, portcheck.StateOpen) {
		t.Fatal("immediate second allow should be suppressed")
	}

	time.Sleep(cooldown + 5*time.Millisecond)

	if !f.Allow(3000, portcheck.StateOpen) {
		t.Fatal("allow after cooldown expiry should pass")
	}
}
