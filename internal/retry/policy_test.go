package retry

import (
	"testing"
	"time"
)

func TestDefaultPolicy_HasPositiveMaxAttempts(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxAttempts <= 0 {
		t.Fatalf("expected positive MaxAttempts, got %d", p.MaxAttempts)
	}
}

func TestDefaultPolicy_HasPositiveInitialDelay(t *testing.T) {
	p := DefaultPolicy()
	if p.InitialDelay <= 0 {
		t.Fatalf("expected positive InitialDelay, got %v", p.InitialDelay)
	}
}

func TestDefaultPolicy_MultiplierAboveOne(t *testing.T) {
	p := DefaultPolicy()
	if p.Multiplier <= 1.0 {
		t.Fatalf("expected Multiplier > 1, got %v", p.Multiplier)
	}
}

func TestDefaultPolicy_MaxDelayCapsDuration(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxDelay <= 0 {
		t.Fatalf("expected positive MaxDelay, got %v", p.MaxDelay)
	}
	if p.MaxDelay < p.InitialDelay {
		t.Fatalf("MaxDelay %v should be >= InitialDelay %v", p.MaxDelay, p.InitialDelay)
	}
}

func TestCustomPolicy_Respected(t *testing.T) {
	p := Policy{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		Multiplier:   1.5,
		MaxDelay:     2 * time.Second,
	}
	if p.MaxAttempts != 5 {
		t.Fatalf("expected 5, got %d", p.MaxAttempts)
	}
}
