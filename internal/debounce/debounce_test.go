package debounce

import (
	"testing"
	"time"
)

func fastPolicy(threshold int) Policy {
	return Policy{Threshold: threshold, Window: 10 * time.Second}
}

func TestObserve_FirstCallNeverEmitsWhenThresholdGtOne(t *testing.T) {
	d := New(fastPolicy(3))
	if d.Observe("host:80", true) {
		t.Fatal("expected false on first observation")
	}
}

func TestObserve_EmitsAfterThresholdReached(t *testing.T) {
	d := New(fastPolicy(3))
	d.Observe("host:80", true)
	d.Observe("host:80", true)
	if !d.Observe("host:80", true) {
		t.Fatal("expected true on third consecutive observation")
	}
}

func TestObserve_CounterResetsOnValueChange(t *testing.T) {
	d := New(fastPolicy(2))
	d.Observe("host:80", true)
	// flip value — counter must reset
	d.Observe("host:80", false)
	if d.Observe("host:80", false) {
		// only 2 consecutive false, but first false resets, so count=2 → emit
		// This is actually expected to emit at count==2 with threshold 2.
		// Rewrite: threshold 3 so the second false is count=2, not yet 3.
	}
	d2 := New(fastPolicy(3))
	d2.Observe("host:80", true)
	d2.Observe("host:80", true)
	d2.Observe("host:80", false) // reset
	if d2.Observe("host:80", false) {
		t.Fatal("count should be 2 after reset, threshold 3 not reached")
	}
}

func TestObserve_WindowExpiryResetsCount(t *testing.T) {
	p := Policy{Threshold: 2, Window: 50 * time.Millisecond}
	d := New(p)

	faketime := time.Now()
	d.now = func() time.Time { return faketime }

	d.Observe("host:80", true)
	// advance past window
	faketime = faketime.Add(100 * time.Millisecond)
	if d.Observe("host:80", true) {
		t.Fatal("count should have reset after window expiry; threshold not met")
	}
}

func TestObserve_DifferentKeysAreIndependent(t *testing.T) {
	d := New(fastPolicy(2))
	d.Observe("host:80", true)
	d.Observe("host:443", true)
	if d.Observe("host:80", true) == false {
		t.Fatal("host:80 should have reached threshold=2")
	}
	if d.Observe("host:443", true) == false {
		t.Fatal("host:443 should have reached threshold=2 independently")
	}
}

func TestReset_ClearsState(t *testing.T) {
	d := New(fastPolicy(2))
	d.Observe("host:80", true)
	d.Reset()
	if d.Observe("host:80", true) {
		t.Fatal("after Reset, count should restart; threshold 2 not met on first call")
	}
}

func TestNew_ClampsInvalidThreshold(t *testing.T) {
	d := New(Policy{Threshold: 0, Window: time.Second})
	// threshold clamped to 1 → first call emits
	if !d.Observe("k", true) {
		t.Fatal("threshold 1 should emit on first call")
	}
}

func TestNew_DefaultPolicyValues(t *testing.T) {
	p := DefaultPolicy()
	if p.Threshold < 1 {
		t.Errorf("default threshold must be >= 1, got %d", p.Threshold)
	}
	if p.Window <= 0 {
		t.Errorf("default window must be positive, got %v", p.Window)
	}
}
