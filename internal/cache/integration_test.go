package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/cache"
)

func TestCache_ConcurrentSafety(t *testing.T) {
	c := cache.New[int, int](time.Hour)
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			c.Set(i, i*10)
		}()
		go func() {
			defer wg.Done()
			c.Get(i)
		}()
	}
	wg.Wait()
}

func TestCache_OverwriteResetsExpiry(t *testing.T) {
	c := cache.New[string, int](30 * time.Millisecond)
	c.Set("port", 1)
	time.Sleep(20 * time.Millisecond)
	// Overwrite before expiry — new TTL should start fresh.
	c.Set("port", 2)
	time.Sleep(20 * time.Millisecond)
	// Total elapsed ~40 ms; original TTL would have expired, new one should not.
	v, ok := c.Get("port")
	if !ok {
		t.Fatal("expected hit after overwrite refreshed TTL")
	}
	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}

func TestCache_PurgeAndContinuedUse(t *testing.T) {
	c := cache.New[string, string](15 * time.Millisecond)
	c.Set("a", "alpha")
	time.Sleep(20 * time.Millisecond)
	c.Purge()

	// Cache should still be usable after purge.
	c.Set("b", "beta")
	v, ok := c.Get("b")
	if !ok || v != "beta" {
		t.Fatal("expected cache to work normally after purge")
	}
}
