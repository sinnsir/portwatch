package backoff_test

import (
	"testing"
	"time"

	"portwatch/internal/backoff"
)

func TestDefaultPolicy_FieldsArePositive(t *testing.T) {
	p := backoff.DefaultPolicy()
	if p.InitialDelay <= 0 {
		t.Errorf("expected positive InitialDelay, got %v", p.InitialDelay)
	}
	if p.Multiplier <= 1 {
		t.Errorf("expected Multiplier > 1, got %v", p.Multiplier)
	}
	if p.MaxDelay <= 0 {
		t.Errorf("expected positive MaxDelay, got %v", p.MaxDelay)
	}
	if p.Jitter < 0 || p.Jitter > 1 {
		t.Errorf("expected Jitter in [0,1], got %v", p.Jitter)
	}
}

func TestDuration_GrowsWithAttempt(t *testing.T) {
	p := backoff.Policy{
		InitialDelay: 100 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     10 * time.Second,
		Jitter:       0, // no jitter for deterministic test
	}

	d0 := p.Duration(0)
	d1 := p.Duration(1)
	d2 := p.Duration(2)

	if d1 <= d0 {
		t.Errorf("expected d1 > d0, got d0=%v d1=%v", d0, d1)
	}
	if d2 <= d1 {
		t.Errorf("expected d2 > d1, got d1=%v d2=%v", d1, d2)
	}
}

func TestDuration_CapsAtMaxDelay(t *testing.T) {
	p := backoff.Policy{
		InitialDelay: 1 * time.Second,
		Multiplier:   10.0,
		MaxDelay:     5 * time.Second,
		Jitter:       0,
	}

	for attempt := 0; attempt < 10; attempt++ {
		d := p.Duration(attempt)
		if d > p.MaxDelay {
			t.Errorf("attempt %d: duration %v exceeds MaxDelay %v", attempt, d, p.MaxDelay)
		}
	}
}

func TestDuration_NegativeAttemptTreatedAsZero(t *testing.T) {
	p := backoff.Policy{
		InitialDelay: 200 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     10 * time.Second,
		Jitter:       0,
	}

	if p.Duration(-1) != p.Duration(0) {
		t.Error("expected Duration(-1) == Duration(0)")
	}
}

func TestDuration_JitterProducesVariance(t *testing.T) {
	p := backoff.Policy{
		InitialDelay: 1 * time.Second,
		Multiplier:   1.0,
		MaxDelay:     10 * time.Second,
		Jitter:       0.5,
	}

	seen := make(map[time.Duration]bool)
	for i := 0; i < 20; i++ {
		seen[p.Duration(0)] = true
	}
	if len(seen) < 2 {
		t.Error("expected jitter to produce varied durations")
	}
}
