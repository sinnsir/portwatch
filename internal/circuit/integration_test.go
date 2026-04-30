package circuit_test

import (
	"sync"
	"testing"
	"time"

	"portwatch/internal/circuit"
)

func TestBreaker_ConcurrentSafety(t *testing.T) {
	b := circuit.New(circuit.Policy{MaxFailures: 10, Cooldown: time.Second})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				b.RecordFailure()
			} else {
				b.RecordSuccess()
			}
			_ = b.Allow()
			_ = b.State()
		}(i)
	}
	wg.Wait()
}

func TestBreaker_ReOpensOnProbeFailure(t *testing.T) {
	p := circuit.Policy{MaxFailures: 2, Cooldown: 30 * time.Millisecond}
	b := circuit.New(p)

	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuit.StateOpen {
		t.Fatal("expected open")
	}

	time.Sleep(p.Cooldown + 10*time.Millisecond)

	// probe allowed
	if err := b.Allow(); err != nil {
		t.Fatalf("probe should pass: %v", err)
	}
	// probe fails — circuit re-opens
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuit.StateOpen {
		t.Fatalf("expected re-open after probe failure, got %s", b.State())
	}
}

func TestBreaker_FullCycle(t *testing.T) {
	p := circuit.Policy{MaxFailures: 2, Cooldown: 20 * time.Millisecond}
	b := circuit.New(p)

	// closed → open
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuit.StateOpen {
		t.Fatal("expected open")
	}

	// wait for cooldown
	time.Sleep(p.Cooldown + 10*time.Millisecond)

	// half-open probe succeeds → closed
	if err := b.Allow(); err != nil {
		t.Fatalf("half-open allow: %v", err)
	}
	b.RecordSuccess()
	if b.State() != circuit.StateClosed {
		t.Fatalf("expected closed after recovery, got %s", b.State())
	}
}
