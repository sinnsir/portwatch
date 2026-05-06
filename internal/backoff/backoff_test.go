package backoff_test

import (
	"testing"
	"time"

	"portwatch/internal/backoff"
)

func noJitterPolicy() backoff.Policy {
	return backoff.Policy{
		InitialDelay: 100 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     1 * time.Second,
		Jitter:       false,
	}
}

func TestNext_FirstCallReturnsZero(t *testing.T) {
	c := backoff.New(noJitterPolicy())
	if d := c.Next(); d != 0 {
		t.Fatalf("expected 0 for first call, got %v", d)
	}
}

func TestNext_SecondCallReturnsInitialDelay(t *testing.T) {
	c := backoff.New(noJitterPolicy())
	c.Next() // attempt 0 → 0
	d := c.Next() // attempt 1 → InitialDelay * 2^0 = 100 ms
	if d != 100*time.Millisecond {
		t.Fatalf("expected 100ms, got %v", d)
	}
}

func TestNext_ExponentialGrowth(t *testing.T) {
	c := backoff.New(noJitterPolicy())
	expected := []time.Duration{0, 100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond}
	for i, want := range expected {
		got := c.Next()
		if got != want {
			t.Fatalf("attempt %d: expected %v, got %v", i, want, got)
		}
	}
}

func TestNext_CapsAtMaxDelay(t *testing.T) {
	c := backoff.New(noJitterPolicy())
	var last time.Duration
	for i := 0; i < 20; i++ {
		last = c.Next()
	}
	if last > 1*time.Second {
		t.Fatalf("delay exceeded MaxDelay: %v", last)
	}
}

func TestReset_RestartsSequence(t *testing.T) {
	c := backoff.New(noJitterPolicy())
	c.Next()
	c.Next()
	c.Reset()
	if c.Attempt() != 0 {
		t.Fatalf("expected attempt 0 after reset, got %d", c.Attempt())
	}
	if d := c.Next(); d != 0 {
		t.Fatalf("expected 0 after reset, got %v", d)
	}
}

func TestDefaultPolicy_HasPositiveInitialDelay(t *testing.T) {
	p := backoff.DefaultPolicy()
	if p.InitialDelay <= 0 {
		t.Fatal("DefaultPolicy InitialDelay must be positive")
	}
}

func TestDefaultPolicy_MultiplierAboveOne(t *testing.T) {
	p := backoff.DefaultPolicy()
	if p.Multiplier <= 1 {
		t.Fatalf("DefaultPolicy Multiplier must be > 1, got %v", p.Multiplier)
	}
}

func TestNew_ZeroMultiplierDefaultsToTwo(t *testing.T) {
	p := backoff.Policy{InitialDelay: 50 * time.Millisecond, Multiplier: 0, MaxDelay: 1 * time.Second}
	c := backoff.New(p)
	c.Next() // skip zero
	d := c.Next()
	if d != 50*time.Millisecond {
		t.Fatalf("expected 50ms with defaulted multiplier, got %v", d)
	}
}
