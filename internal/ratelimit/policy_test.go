package ratelimit

import (
	"testing"
	"time"
)

func TestNew_DefaultWindow(t *testing.T) {
	l := New(5, 0)
	if l.window != time.Minute {
		t.Fatalf("expected default window of 1m, got %s", l.window)
	}
}

func TestNew_CustomRateAndWindow(t *testing.T) {
	l := New(10, 30*time.Second)
	if l.rate != 10 {
		t.Fatalf("expected rate 10, got %d", l.rate)
	}
	if l.window != 30*time.Second {
		t.Fatalf("expected window 30s, got %s", l.window)
	}
}

func TestNew_NegativeRateDefaultsToOne(t *testing.T) {
	l := New(-5, time.Minute)
	if l.rate != 1 {
		t.Fatalf("expected rate clamped to 1, got %d", l.rate)
	}
}

func TestRemaining_UnknownKeyReturnsFullRate(t *testing.T) {
	l := New(7, time.Minute)
	if got := l.Remaining("never-seen"); got != 7 {
		t.Fatalf("expected 7 for unknown key, got %d", got)
	}
}

func TestRemaining_AfterWindowExpiry(t *testing.T) {
	l := New(3, 40*time.Millisecond)
	key := "expiry-remaining"
	l.Allow(key)
	l.Allow(key)

	time.Sleep(50 * time.Millisecond)

	if got := l.Remaining(key); got != 3 {
		t.Fatalf("expected full bucket after window expiry, got %d", got)
	}
}

func TestAllow_DecreasesRemaining(t *testing.T) {
	l := New(5, time.Minute)
	key := "allow-decrements"

	for i := 5; i >= 0; i-- {
		if got := l.Remaining(key); got != i {
			t.Fatalf("expected remaining %d before Allow, got %d", i, got)
		}
		if i > 0 {
			l.Allow(key)
		}
	}
}
