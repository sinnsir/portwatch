package aggregator

import (
	"sync"
	"testing"
	"time"
)

func fastPolicy() Policy {
	return Policy{Window: 20 * time.Millisecond, Capacity: 5}
}

func TestAggregator_FlushOnWindow(t *testing.T) {
	var mu sync.Mutex
	var got []Summary

	a := New(fastPolicy(), func(s Summary) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	})

	a.Add(Event{Host: "localhost", Port: 8080, State: "open"})
	a.Add(Event{Host: "localhost", Port: 9090, State: "closed"})

	time.Sleep(60 * time.Millisecond)
	a.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("expected at least one summary, got none")
	}
	total := 0
	for _, s := range got {
		total += len(s.Events)
	}
	if total != 2 {
		t.Fatalf("expected 2 events total, got %d", total)
	}
}

func TestAggregator_EarlyFlushOnCapacity(t *testing.T) {
	flushCh := make(chan Summary, 10)
	p := Policy{Window: 10 * time.Second, Capacity: 3}

	a := New(p, func(s Summary) { flushCh <- s })
	defer a.Stop()

	for i := 0; i < 3; i++ {
		a.Add(Event{Host: "h", Port: i, State: "open"})
	}

	select {
	case s := <-flushCh:
		if len(s.Events) != 3 {
			t.Fatalf("expected 3 events in early flush, got %d", len(s.Events))
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for early flush")
	}
}

func TestAggregator_StopFlushesRemaining(t *testing.T) {
	var mu sync.Mutex
	var got []Summary
	p := Policy{Window: 10 * time.Second, Capacity: 100}

	a := New(p, func(s Summary) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	})

	a.Add(Event{Host: "localhost", Port: 1234, State: "open"})
	a.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("expected flush on Stop, got none")
	}
	if got[0].Events[0].Port != 1234 {
		t.Fatalf("unexpected port %d", got[0].Events[0].Port)
	}
}

func TestAggregator_EmptyFlushNotCalled(t *testing.T) {
	called := 0
	a := New(fastPolicy(), func(_ Summary) { called++ })
	a.Stop()
	time.Sleep(30 * time.Millisecond)
	if called != 0 {
		t.Fatalf("handler should not be called for empty buffer, got %d calls", called)
	}
}

func TestAggregator_TimestampAutoSet(t *testing.T) {
	flushCh := make(chan Summary, 1)
	p := Policy{Window: 10 * time.Second, Capacity: 1}

	a := New(p, func(s Summary) { flushCh <- s })
	defer a.Stop()

	before := time.Now()
	a.Add(Event{Host: "h", Port: 80, State: "open"}) // At is zero

	select {
	case s := <-flushCh:
		if s.Events[0].At.Before(before) {
			t.Fatal("auto-set timestamp should be >= time before Add")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out")
	}
}
