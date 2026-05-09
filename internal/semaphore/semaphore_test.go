package semaphore_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/semaphore"
)

func TestNew_ClampsToOne(t *testing.T) {
	s := semaphore.New(0)
	if s.Cap() != 1 {
		t.Fatalf("expected cap 1, got %d", s.Cap())
	}
}

func TestAcquire_DecreasesAvailable(t *testing.T) {
	s := semaphore.New(3)
	if err := s.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := s.Available(); got != 2 {
		t.Fatalf("expected 2 available, got %d", got)
	}
	s.Release()
}

func TestRelease_RestoresSlot(t *testing.T) {
	s := semaphore.New(1)
	_ = s.Acquire(context.Background())
	s.Release()
	if s.Available() != 1 {
		t.Fatal("slot not restored after release")
	}
}

func TestTryAcquire_SucceedsWhenAvailable(t *testing.T) {
	s := semaphore.New(1)
	if !s.TryAcquire() {
		t.Fatal("expected TryAcquire to succeed")
	}
	if s.TryAcquire() {
		t.Fatal("expected TryAcquire to fail when exhausted")
	}
	s.Release()
}

func TestAcquire_CancelledContext(t *testing.T) {
	s := semaphore.New(1)
	_ = s.Acquire(context.Background()) // exhaust

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := s.Acquire(ctx); err == nil {
		t.Fatal("expected ErrAcquire, got nil")
	}
	s.Release()
}

func TestAcquire_ConcurrentSafety(t *testing.T) {
	const (
		cap     = 4
		workers = 32
	)
	s := semaphore.New(cap)
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		peak    int
		current int
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Acquire(context.Background())
			mu.Lock()
			current++
			if current > peak {
				peak = current
			}
			mu.Unlock()
			time.Sleep(5 * time.Millisecond)
			mu.Lock()
			current--
			mu.Unlock()
			s.Release()
		}()
	}
	wg.Wait()
	if peak > cap {
		t.Fatalf("concurrency exceeded cap: peak=%d cap=%d", peak, cap)
	}
}
