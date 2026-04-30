package circuit

import (
	"testing"
	"time"
)

func TestDefaultPolicy_HasPositiveMaxFailures(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxFailures <= 0 {
		t.Fatalf("expected positive MaxFailures, got %d", p.MaxFailures)
	}
}

func TestDefaultPolicy_HasPositiveCooldown(t *testing.T) {
	p := DefaultPolicy()
	if p.Cooldown <= 0 {
		t.Fatalf("expected positive Cooldown, got %v", p.Cooldown)
	}
}

func TestDefaultPolicy_CooldownAtLeastOneSecond(t *testing.T) {
	p := DefaultPolicy()
	if p.Cooldown < time.Second {
		t.Fatalf("cooldown %v is less than 1s", p.Cooldown)
	}
}

func TestNew_NegativeMaxFailuresDefaultsToOne(t *testing.T) {
	b := New(Policy{MaxFailures: -1, Cooldown: time.Second})
	// Should not panic; internal cap applied.
	if b == nil {
		t.Fatal("expected non-nil breaker")
	}
	// One failure should open it.
	b.RecordFailure()
	if b.State() != StateOpen {
		t.Fatalf("expected open after 1 failure with cap, got %s", b.State())
	}
}

func TestNew_ZeroCooldownDefaultsToOneSecond(t *testing.T) {
	b := New(Policy{MaxFailures: 1, Cooldown: 0})
	if b == nil {
		t.Fatal("expected non-nil breaker")
	}
	b.RecordFailure()
	// Should be open; cooldown is 1s so still open immediately.
	if err := b.Allow(); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}
