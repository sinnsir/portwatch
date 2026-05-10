package cooldown

import (
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{Window: 50 * time.Millisecond}
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	tr := New(fastPolicy())
	if !tr.Allow("port:8080") {
		t.Fatal("expected first call to pass")
	}
}

func TestAllow_SecondCallSuppressedWithinWindow(t *testing.T) {
	tr := New(fastPolicy())
	tr.Allow("port:8080")
	if tr.Allow("port:8080") {
		t.Fatal("expected second call within window to be suppressed")
	}
}

func TestAllow_PassesAfterWindowExpires(t *testing.T) {
	tr := New(fastPolicy())
	tr.Allow("port:8080")
	time.Sleep(60 * time.Millisecond)
	if !tr.Allow("port:8080") {
		t.Fatal("expected call after window expiry to pass")
	}
}

func TestAllow_DifferentKeysAreIndependent(t *testing.T) {
	tr := New(fastPolicy())
	tr.Allow("port:8080")
	if !tr.Allow("port:9090") {
		t.Fatal("expected different key to pass independently")
	}
}

func TestReset_AllowsImmediateReentry(t *testing.T) {
	tr := New(fastPolicy())
	tr.Allow("port:8080")
	tr.Reset("port:8080")
	if !tr.Allow("port:8080") {
		t.Fatal("expected reset key to pass immediately")
	}
}

func TestRemaining_ZeroForUnknownKey(t *testing.T) {
	tr := New(fastPolicy())
	if r := tr.Remaining("port:8080"); r != 0 {
		t.Fatalf("expected 0, got %v", r)
	}
}

func TestRemaining_PositiveWhileInCooldown(t *testing.T) {
	tr := New(fastPolicy())
	tr.Allow("port:8080")
	if r := tr.Remaining("port:8080"); r <= 0 {
		t.Fatalf("expected positive remaining, got %v", r)
	}
}

func TestRemaining_ZeroAfterWindowExpires(t *testing.T) {
	tr := New(fastPolicy())
	tr.Allow("port:8080")
	time.Sleep(60 * time.Millisecond)
	if r := tr.Remaining("port:8080"); r != 0 {
		t.Fatalf("expected 0 after expiry, got %v", r)
	}
}

func TestDefaultPolicy_HasPositiveWindow(t *testing.T) {
	p := DefaultPolicy()
	if p.Window <= 0 {
		t.Fatalf("expected positive window, got %v", p.Window)
	}
}

func TestDefaultPolicy_WindowAtLeastOneSecond(t *testing.T) {
	p := DefaultPolicy()
	if p.Window < time.Second {
		t.Fatalf("expected window >= 1s, got %v", p.Window)
	}
}

func TestZeroWindow_DefaultsToPositive(t *testing.T) {
	tr := New(Policy{Window: 0})
	tr.Allow("k")
	// second call must be suppressed — window defaulted to a positive value
	if tr.Allow("k") {
		t.Fatal("expected suppression even with zero-window policy (defaulted)")
	}
}
