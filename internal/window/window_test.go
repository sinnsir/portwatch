package window

import (
	"sync"
	"testing"
	"time"
)

func fastPolicy(size time.Duration) Policy {
	return Policy{Size: size}
}

func TestAdd_IncrementsCount(t *testing.T) {
	w := New(fastPolicy(5 * time.Second))

	got := w.Add("k", 3)
	if got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
	got = w.Add("k", 2)
	if got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestCount_ReturnsZeroForUnknownKey(t *testing.T) {
	w := New(fastPolicy(5 * time.Second))
	if c := w.Count("missing"); c != 0 {
		t.Fatalf("expected 0, got %d", c)
	}
}

func TestAdd_DifferentKeysIndependent(t *testing.T) {
	w := New(fastPolicy(5 * time.Second))
	w.Add("a", 10)
	w.Add("b", 1)

	if c := w.Count("a"); c != 10 {
		t.Fatalf("a: expected 10, got %d", c)
	}
	if c := w.Count("b"); c != 1 {
		t.Fatalf("b: expected 1, got %d", c)
	}
}

func TestAdd_ExpiredBucketsNotCounted(t *testing.T) {
	w := New(fastPolicy(100 * time.Millisecond))

	base := time.Now()
	w.now = func() time.Time { return base }
	w.Add("k", 7)

	// Advance past the window.
	w.now = func() time.Time { return base.Add(200 * time.Millisecond) }
	if c := w.Count("k"); c != 0 {
		t.Fatalf("expected 0 after expiry, got %d", c)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	w := New(fastPolicy(5 * time.Second))
	w.Add("k", 99)
	w.Reset("k")
	if c := w.Count("k"); c != 0 {
		t.Fatalf("expected 0 after reset, got %d", c)
	}
}

func TestDefaultPolicy_HasPositiveSize(t *testing.T) {
	p := DefaultPolicy()
	if p.Size <= 0 {
		t.Fatalf("expected positive Size, got %v", p.Size)
	}
}

func TestNew_ZeroSizeDefaulted(t *testing.T) {
	w := New(Policy{Size: 0})
	if w.policy.Size <= 0 {
		t.Fatalf("expected positive Size after defaulting, got %v", w.policy.Size)
	}
}

func TestAdd_ConcurrentSafety(t *testing.T) {
	w := New(fastPolicy(5 * time.Second))
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w.Add("shared", 1)
		}()
	}
	wg.Wait()
	if c := w.Count("shared"); c != 50 {
		t.Fatalf("expected 50, got %d", c)
	}
}
