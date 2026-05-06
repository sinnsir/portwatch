package sampler

import (
	"testing"
)

func TestAllow_RateOne_AlwaysPasses(t *testing.T) {
	s := New(Policy{Rate: 1.0})
	for i := 0; i < 100; i++ {
		if !s.Allow() {
			t.Fatal("expected Allow to return true for rate=1.0")
		}
	}
}

func TestAllow_RateZero_DefaultsToOne(t *testing.T) {
	s := New(Policy{Rate: 0})
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate to be clamped to 1.0, got %v", s.Rate())
	}
	for i := 0; i < 50; i++ {
		if !s.Allow() {
			t.Fatal("expected Allow to return true after clamping")
		}
	}
}

func TestAllow_NegativeRate_DefaultsToOne(t *testing.T) {
	s := New(Policy{Rate: -0.5})
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate clamped to 1.0, got %v", s.Rate())
	}
}

func TestAllow_LowRate_ReducesThroughput(t *testing.T) {
	s := New(Policy{Rate: 0.1})
	passed := 0
	const trials = 10_000
	for i := 0; i < trials; i++ {
		if s.Allow() {
			passed++
		}
	}
	// Expect roughly 10 % ± 5 % to pass.
	lo, hi := int(trials*0.05), int(trials*0.15)
	if passed < lo || passed > hi {
		t.Fatalf("expected %d–%d events to pass, got %d", lo, hi, passed)
	}
}

func TestRate_ReturnsConfiguredValue(t *testing.T) {
	s := New(Policy{Rate: 0.42})
	if s.Rate() != 0.42 {
		t.Fatalf("expected 0.42, got %v", s.Rate())
	}
}

func TestDefaultPolicy_RateIsOne(t *testing.T) {
	p := DefaultPolicy()
	if p.Rate != 1.0 {
		t.Fatalf("expected DefaultPolicy Rate=1.0, got %v", p.Rate)
	}
}
