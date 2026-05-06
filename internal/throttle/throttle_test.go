package throttle

import (
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{
		Rate:           3,
		RefillInterval: 50 * time.Millisecond,
	}
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	th := New(fastPolicy())
	if !th.Allow() {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_ExhaustBucket(t *testing.T) {
	th := New(fastPolicy()) // rate=3
	for i := 0; i < 3; i++ {
		if !th.Allow() {
			t.Fatalf("call %d should be allowed", i)
		}
	}
	if th.Allow() {
		t.Fatal("expected 4th call to be denied")
	}
}

func TestAllow_RefillAddsTokens(t *testing.T) {
	now := time.Now()
	th := New(fastPolicy())
	th.now = func() time.Time { return now }

	// drain
	for i := 0; i < 3; i++ {
		th.Allow()
	}
	if th.Allow() {
		t.Fatal("should be empty")
	}

	// advance time by 2 refill intervals → +2 tokens
	th.now = func() time.Time { return now.Add(100 * time.Millisecond) }
	if !th.Allow() {
		t.Fatal("should have tokens after refill")
	}
	if !th.Allow() {
		t.Fatal("second refilled token should be available")
	}
	if th.Allow() {
		t.Fatal("third call should be denied after using refilled tokens")
	}
}

func TestAllow_DoesNotExceedRate(t *testing.T) {
	now := time.Now()
	th := New(fastPolicy()) // rate=3
	th.now = func() time.Time { return now.Add(10 * time.Second) } // huge advance

	count := 0
	for i := 0; i < 10; i++ {
		if th.Allow() {
			count++
		}
	}
	if count > 3 {
		t.Fatalf("tokens should be capped at rate=3, got %d", count)
	}
}

func TestRemaining_CorrectCount(t *testing.T) {
	th := New(fastPolicy()) // rate=3
	if th.Remaining() != 3 {
		t.Fatalf("expected 3, got %d", th.Remaining())
	}
	th.Allow()
	if th.Remaining() != 2 {
		t.Fatalf("expected 2, got %d", th.Remaining())
	}
}

func TestNew_ZeroRateDefaultsToSensible(t *testing.T) {
	th := New(Policy{})
	if th.policy.Rate <= 0 {
		t.Fatal("expected positive rate from default policy")
	}
	if th.policy.RefillInterval <= 0 {
		t.Fatal("expected positive refill interval from default policy")
	}
}
