package cache_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/cache"
)

func TestSet_And_Get_HitBeforeExpiry(t *testing.T) {
	c := cache.New[string, int](time.Hour)
	c.Set("key", 42)
	v, ok := c.Get("key")
	if !ok {
		t.Fatal("expected hit, got miss")
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestGet_MissOnUnknownKey(t *testing.T) {
	c := cache.New[string, string](time.Hour)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected miss for unknown key")
	}
}

func TestGet_MissAfterExpiry(t *testing.T) {
	c := cache.New[string, int](10 * time.Millisecond)
	c.Set("x", 7)
	time.Sleep(20 * time.Millisecond)
	_, ok := c.Get("x")
	if ok {
		t.Fatal("expected miss after TTL expiry")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	c := cache.New[string, bool](time.Hour)
	c.Set("flag", true)
	c.Delete("flag")
	_, ok := c.Get("flag")
	if ok {
		t.Fatal("expected miss after delete")
	}
}

func TestPurge_EvictsExpired(t *testing.T) {
	c := cache.New[string, int](10 * time.Millisecond)
	c.Set("a", 1)
	c.Set("b", 2)
	time.Sleep(20 * time.Millisecond)
	c.Set("c", 3) // fresh entry

	evicted := c.Purge()
	if evicted != 2 {
		t.Fatalf("expected 2 evictions, got %d", evicted)
	}
	if c.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", c.Len())
	}
}

func TestLen_ReflectsInserts(t *testing.T) {
	c := cache.New[int, string](time.Hour)
	if c.Len() != 0 {
		t.Fatal("expected empty cache")
	}
	c.Set(1, "one")
	c.Set(2, "two")
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}

func TestNew_ZeroTTL_DefaultsToThirtySeconds(t *testing.T) {
	c := cache.New[string, int](0)
	c.Set("k", 99)
	v, ok := c.Get("k")
	if !ok || v != 99 {
		t.Fatal("expected hit with default TTL")
	}
}
