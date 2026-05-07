package buffer_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/buffer"
)

func TestBuffer_ConcurrentAdd(t *testing.T) {
	var total atomic.Int64
	p := buffer.Policy{Capacity: 10, FlushInterval: 50 * time.Millisecond}
	b := buffer.New[int](p, func(items []int) {
		total.Add(int64(len(items)))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	go b.Run(ctx)

	const goroutines = 20
	const perGoroutine = 25
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				b.Add(j)
			}
		}()
	}
	wg.Wait()
	<-ctx.Done()
	time.Sleep(30 * time.Millisecond) // let final flush complete

	want := int64(goroutines * perGoroutine)
	if got := total.Load(); got != want {
		t.Errorf("expected %d total items flushed, got %d", want, got)
	}
}

func TestBuffer_MultipleFlushBatches(t *testing.T) {
	var mu sync.Mutex
	var batches [][]int
	p := buffer.Policy{Capacity: 3, FlushInterval: 10 * time.Second}
	b := buffer.New[int](p, func(items []int) {
		mu.Lock()
		batches = append(batches, append([]int(nil), items...))
		mu.Unlock()
	})
	for i := 0; i < 9; i++ {
		b.Add(i)
	}
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches of 3, got %d batches", len(batches))
	}
	for i, batch := range batches {
		if len(batch) != 3 {
			t.Errorf("batch %d: expected 3 items, got %d", i, len(batch))
		}
	}
}
