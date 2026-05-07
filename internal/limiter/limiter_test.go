package limiter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/limiter"
)

func fastPolicy() limiter.Policy {
	return limiter.Policy{Max: 3, Window: 100 * time.Millisecond}
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := limiter.New(fastPolicy())
	if !l.Allow("k") {
		t.Fatal("expected first call to be permitted")
	}
}

func TestAllow_ExhaustLimit(t *testing.T) {
	l := limiter.New(fastPolicy())
	for i := 0; i < 3; i++ {
		if !l.Allow("k") {
			t.Fatalf("call %d should be permitted", i+1)
		}
	}
	if l.Allow("k") {
		t.Fatal("expected call beyond limit to be denied")
	}
}

func TestAllow_DifferentKeysIndependent(t *testing.T) {
	l := limiter.New(fastPolicy())
	for i := 0; i < 3; i++ {
		l.Allow("a")
	}
	if !l.Allow("b") {
		t.Fatal("key b should not be affected by key a's usage")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	l := limiter.New(fastPolicy())
	for i := 0; i < 3; i++ {
		l.Allow("k")
	}
	time.Sleep(120 * time.Millisecond)
	if !l.Allow("k") {
		t.Fatal("expected call to be permitted after window expiry")
	}
}

func TestRemaining_FullOnNewKey(t *testing.T) {
	l := limiter.New(fastPolicy())
	if got := l.Remaining("new"); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
}

func TestRemaining_DecreasesAfterAllow(t *testing.T) {
	l := limiter.New(fastPolicy())
	l.Allow("k")
	l.Allow("k")
	if got := l.Remaining("k"); got != 1 {
		t.Fatalf("expected 1 remaining, got %d", got)
	}
}

func TestAllow_DefaultPolicyAppliedOnZeroValues(t *testing.T) {
	l := limiter.New(limiter.Policy{})
	def := limiter.DefaultPolicy()
	for i := 0; i < def.Max; i++ {
		if !l.Allow("k") {
			t.Fatalf("call %d should be permitted under default policy", i+1)
		}
	}
	if l.Allow("k") {
		t.Fatal("expected call beyond default limit to be denied")
	}
}

func TestAllow_ConcurrentSafety(t *testing.T) {
	l := limiter.New(limiter.Policy{Max: 100, Window: time.Second})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Allow("shared")
			l.Remaining("shared")
		}()
	}
	wg.Wait()
}
