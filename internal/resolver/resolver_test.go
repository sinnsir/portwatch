package resolver_test

import (
	"sync"
	"testing"

	"github.com/user/portwatch/internal/resolver"
)

func TestLookup_KnownEntry(t *testing.T) {
	r := resolver.New([]resolver.Entry{
		{Host: "localhost", Port: 5432, Service: "postgres"},
	})
	svc, ok := r.Lookup("localhost", 5432)
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if svc != "postgres" {
		t.Fatalf("expected postgres, got %q", svc)
	}
}

func TestLookup_UnknownEntry(t *testing.T) {
	r := resolver.New(nil)
	_, ok := r.Lookup("localhost", 9999)
	if ok {
		t.Fatal("expected miss for unknown entry")
	}
}

func TestRegister_AddsEntry(t *testing.T) {
	r := resolver.New(nil)
	r.Register("127.0.0.1", 6379, "redis")
	svc, ok := r.Lookup("127.0.0.1", 6379)
	if !ok || svc != "redis" {
		t.Fatalf("expected redis, got %q ok=%v", svc, ok)
	}
}

func TestRegister_OverwritesExisting(t *testing.T) {
	r := resolver.New([]resolver.Entry{
		{Host: "host", Port: 80, Service: "old"},
	})
	r.Register("host", 80, "new")
	svc, _ := r.Lookup("host", 80)
	if svc != "new" {
		t.Fatalf("expected new, got %q", svc)
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	r := resolver.New([]resolver.Entry{
		{Host: "host", Port: 443, Service: "https"},
	})
	r.Remove("host", 443)
	_, ok := r.Lookup("host", 443)
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestRemove_NoopOnMissing(t *testing.T) {
	r := resolver.New(nil)
	r.Remove("ghost", 1234) // must not panic
}

func TestLen_ReflectsRegistrations(t *testing.T) {
	r := resolver.New(nil)
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
	r.Register("a", 1, "svc-a")
	r.Register("b", 2, "svc-b")
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
	r.Remove("a", 1)
	if r.Len() != 1 {
		t.Fatalf("expected 1 after remove, got %d", r.Len())
	}
}

func TestResolver_ConcurrentSafety(t *testing.T) {
	r := resolver.New(nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			r.Register("host", p, "svc")
			r.Lookup("host", p)
			r.Remove("host", p)
		}(i)
	}
	wg.Wait()
}
