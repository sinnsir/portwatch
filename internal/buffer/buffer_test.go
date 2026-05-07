package buffer

import (
	"context"
	"sync"
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{Capacity: 4, FlushInterval: 50 * time.Millisecond}
}

func TestAdd_CapacityFlush(t *testing.T) {
	var mu sync.Mutex
	var got [][]int
	b := New[int](fastPolicy(), func(items []int) {
		mu.Lock()
		got = append(got, items)
		mu.Unlock()
	})
	for i := 0; i < 4; i++ {
		b.Add(i)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("expected 1 flush batch, got %d", len(got))
	}
	if len(got[0]) != 4 {
		t.Fatalf("expected 4 items in batch, got %d", len(got[0]))
	}
}

func TestAdd_TimerFlush(t *testing.T) {
	var mu sync.Mutex
	var got [][]int
	p := Policy{Capacity: 100, FlushInterval: 40 * time.Millisecond}
	b := New[int](p, func(items []int) {
		mu.Lock()
		got = append(got, items)
		mu.Unlock()
	})
	b.Add(1)
	b.Add(2)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go b.Run(ctx)
	<-ctx.Done()
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("expected at least one timer-triggered flush")
	}
}

func TestFlush_EmptyBufferNoCall(t *testing.T) {
	called := false
	b := New[string](fastPolicy(), func(_ []string) { called = true })
	b.Flush()
	if called {
		t.Fatal("flush callback should not be called on empty buffer")
	}
}

func TestStop_FlushesPending(t *testing.T) {
	var mu sync.Mutex
	var got [][]string
	b := New[string](Policy{Capacity: 100, FlushInterval: 10 * time.Second}, func(items []string) {
		mu.Lock()
		got = append(got, items)
		mu.Unlock()
	})
	b.Add("a")
	b.Add("b")
	go b.Run(context.Background())
	time.Sleep(10 * time.Millisecond)
	b.Stop()
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("Stop should have flushed pending items")
	}
}

func TestDefaultPolicy_SaneValues(t *testing.T) {
	p := DefaultPolicy()
	if p.Capacity <= 0 {
		t.Errorf("capacity must be positive, got %d", p.Capacity)
	}
	if p.FlushInterval <= 0 {
		t.Errorf("flush interval must be positive, got %v", p.FlushInterval)
	}
}
