package ratelimit

import (
	"testing"
	"time"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	l := New(3, time.Minute)
	if !l.Allow("port:8080") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_ExhaustBucket(t *testing.T) {
	l := New(2, time.Minute)
	key := "port:9090"

	if !l.Allow(key) {
		t.Fatal("call 1 should be allowed")
	}
	if !l.Allow(key) {
		t.Fatal("call 2 should be allowed")
	}
	if l.Allow(key) {
		t.Fatal("call 3 should be denied (bucket exhausted)")
	}
}

func TestAllow_DifferentKeysIndependent(t *testing.T) {
	l := New(1, time.Minute)

	if !l.Allow("a") {
		t.Fatal("key a first call should pass")
	}
	if l.Allow("a") {
		t.Fatal("key a second call should be denied")
	}
	if !l.Allow("b") {
		t.Fatal("key b first call should pass independently")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	l := New(1, 50*time.Millisecond)
	key := "port:1234"

	if !l.Allow(key) {
		t.Fatal("first call should pass")
	}
	if l.Allow(key) {
		t.Fatal("second call should be denied")
	}

	time.Sleep(60 * time.Millisecond)

	if !l.Allow(key) {
		t.Fatal("call after window expiry should pass")
	}
}

func TestRemaining_CorrectCount(t *testing.T) {
	l := New(3, time.Minute)
	key := "port:7777"

	if got := l.Remaining(key); got != 3 {
		t.Fatalf("expected 3 remaining before any calls, got %d", got)
	}
	l.Allow(key)
	if got := l.Remaining(key); got != 2 {
		t.Fatalf("expected 2 remaining after one allow, got %d", got)
	}
}

func TestReset_ClearsBucket(t *testing.T) {
	l := New(1, time.Minute)
	key := "port:5555"

	l.Allow(key)
	if l.Allow(key) {
		t.Fatal("should be denied before reset")
	}
	l.Reset(key)
	if !l.Allow(key) {
		t.Fatal("should be allowed after reset")
	}
}

func TestNew_ZeroRateDefaultsToOne(t *testing.T) {
	l := New(0, time.Minute)
	key := "k"
	if !l.Allow(key) {
		t.Fatal("first call should pass even with zero-rate input")
	}
	if l.Allow(key) {
		t.Fatal("second call should be denied with effective rate=1")
	}
}
