package dispatcher

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/monitor"
	"github.com/user/portwatch/internal/portcheck"
)

func makeEvent(port int) monitor.Event {
	return monitor.Event{
		Host:  "localhost",
		Port:  port,
		State: portcheck.StateOpen,
	}
}

func TestDispatch_CallsAllHandlers(t *testing.T) {
	d := New()
	var countA, countB atomic.Int32
	d.Register(func(_ context.Context, _ monitor.Event) { countA.Add(1) })
	d.Register(func(_ context.Context, _ monitor.Event) { countB.Add(1) })

	ch := make(chan monitor.Event, 1)
	ch <- makeEvent(8080)
	close(ch)

	d.Dispatch(context.Background(), ch)

	if countA.Load() != 1 || countB.Load() != 1 {
		t.Fatalf("expected each handler called once, got A=%d B=%d", countA.Load(), countB.Load())
	}
}

func TestDispatch_StopsOnContextCancel(t *testing.T) {
	d := New()
	var count atomic.Int32
	d.Register(func(_ context.Context, _ monitor.Event) { count.Add(1) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	ch := make(chan monitor.Event, 3)
	for i := 0; i < 3; i++ {
		ch <- makeEvent(8080 + i)
	}

	d.Dispatch(ctx, ch)

	if count.Load() != 0 {
		t.Fatalf("expected no handler calls after cancel, got %d", count.Load())
	}
}

func TestDispatch_MultipleEvents(t *testing.T) {
	d := New()
	var count atomic.Int32
	d.Register(func(_ context.Context, _ monitor.Event) { count.Add(1) })

	ch := make(chan monitor.Event, 5)
	for i := 0; i < 5; i++ {
		ch <- makeEvent(9000 + i)
	}
	close(ch)

	d.Dispatch(context.Background(), ch)

	if count.Load() != 5 {
		t.Fatalf("expected 5 handler calls, got %d", count.Load())
	}
}

func TestDispatch_ConcurrentRegister(t *testing.T) {
	d := New()
	var mu sync.Mutex
	seen := map[int]int{}

	const n = 10
	for i := 0; i < n; i++ {
		port := 7000 + i
		d.Register(func(_ context.Context, ev monitor.Event) {
			mu.Lock()
			seen[ev.Port]++
			mu.Unlock()
		})
	}

	ch := make(chan monitor.Event, 1)
	ch <- makeEvent(7777)
	close(ch)

	d.Dispatch(context.Background(), ch)

	mu.Lock()
	defer mu.Unlock()
	if seen[7777] != n {
		t.Fatalf("expected %d handler calls, got %d", n, seen[7777])
	}
}

func TestDispatch_NoHandlers(t *testing.T) {
	d := New()
	ch := make(chan monitor.Event, 1)
	ch <- makeEvent(1234)
	close(ch)

	done := make(chan struct{})
	go func() {
		d.Dispatch(context.Background(), ch)
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Dispatch did not return with no handlers")
	}
}
