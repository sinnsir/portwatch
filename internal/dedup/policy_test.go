package dedup

import (
	"testing"
	"time"
)

func TestDefaultPolicy_HasPositiveWindow(t *testing.T) {
	if DefaultPolicy.Window <= 0 {
		t.Fatalf("DefaultPolicy.Window must be positive, got %v", DefaultPolicy.Window)
	}
}

func TestDefaultPolicy_WindowAtLeastOneSecond(t *testing.T) {
	if DefaultPolicy.Window < time.Second {
		t.Fatalf("DefaultPolicy.Window should be >= 1s, got %v", DefaultPolicy.Window)
	}
}

func TestCustomPolicy_Respected(t *testing.T) {
	p := Policy{Window: 5 * time.Minute}
	d := New(p)
	if d.window != 5*time.Minute {
		t.Fatalf("expected 5m window, got %v", d.window)
	}
}

func TestNegativeWindow_DefaultsToPositive(t *testing.T) {
	d := New(Policy{Window: -1 * time.Second})
	if d.window <= 0 {
		t.Fatal("negative window should fall back to a positive default")
	}
}
