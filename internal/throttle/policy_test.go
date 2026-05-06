package throttle

import (
	"testing"
	"time"
)

func TestDefaultPolicy_HasPositiveRate(t *testing.T) {
	p := DefaultPolicy()
	if p.Rate <= 0 {
		t.Fatalf("expected positive Rate, got %d", p.Rate)
	}
}

func TestDefaultPolicy_HasPositiveRefillInterval(t *testing.T) {
	p := DefaultPolicy()
	if p.RefillInterval <= 0 {
		t.Fatalf("expected positive RefillInterval, got %v", p.RefillInterval)
	}
}

func TestDefaultPolicy_RefillIntervalAtLeast100ms(t *testing.T) {
	p := DefaultPolicy()
	if p.RefillInterval < 100*time.Millisecond {
		t.Fatalf("RefillInterval too aggressive: %v", p.RefillInterval)
	}
}

func TestCustomPolicy_Respected(t *testing.T) {
	p := Policy{Rate: 5, RefillInterval: 200 * time.Millisecond}
	th := New(p)
	if th.policy.Rate != 5 {
		t.Fatalf("expected rate 5, got %d", th.policy.Rate)
	}
	if th.policy.RefillInterval != 200*time.Millisecond {
		t.Fatalf("expected 200ms, got %v", th.policy.RefillInterval)
	}
}
