package dedup

import (
	"fmt"
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{Window: 50 * time.Millisecond}
}

func TestIsDuplicate_FirstCallReturnsFalse(t *testing.T) {
	d := New(fastPolicy())
	if d.IsDuplicate("host:8080:open") {
		t.Fatal("expected false on first call")
	}
}

func TestIsDuplicate_SecondCallReturnsTrue(t *testing.T) {
	d := New(fastPolicy())
	d.IsDuplicate("host:8080:open")
	if !d.IsDuplicate("host:8080:open") {
		t.Fatal("expected duplicate on second call within window")
	}
}

func TestIsDuplicate_DifferentKeysIndependent(t *testing.T) {
	d := New(fastPolicy())
	d.IsDuplicate("host:8080:open")
	if d.IsDuplicate("host:9090:open") {
		t.Fatal("different key should not be considered duplicate")
	}
}

func TestIsDuplicate_ExpiredKeyAllowsReentry(t *testing.T) {
	d := New(fastPolicy())
	d.IsDuplicate("host:8080:open")
	time.Sleep(60 * time.Millisecond)
	if d.IsDuplicate("host:8080:open") {
		t.Fatal("expired key should not be treated as duplicate")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	d := New(fastPolicy())
	for i := 0; i < 5; i++ {
		d.IsDuplicate(fmt.Sprintf("host:%d:open", i))
	}
	if d.Len() != 5 {
		t.Fatalf("expected 5 entries, got %d", d.Len())
	}
	time.Sleep(60 * time.Millisecond)
	d.Purge()
	if d.Len() != 0 {
		t.Fatalf("expected 0 entries after purge, got %d", d.Len())
	}
}

func TestNew_ZeroWindowUsesDefault(t *testing.T) {
	d := New(Policy{Window: 0})
	if d.window != DefaultPolicy.Window {
		t.Fatalf("expected default window %v, got %v", DefaultPolicy.Window, d.window)
	}
}

func TestLen_ReflectsTrackedKeys(t *testing.T) {
	d := New(fastPolicy())
	if d.Len() != 0 {
		t.Fatal("expected empty deduplicator")
	}
	d.IsDuplicate("a")
	d.IsDuplicate("b")
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
}
