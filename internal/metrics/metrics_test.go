package metrics

import (
	"sync"
	"testing"
)

func TestCounter_IncAndValue(t *testing.T) {
	var c Counter
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %d", c.Value())
	}
	c.Inc()
	c.Inc()
	if c.Value() != 2 {
		t.Fatalf("expected 2, got %d", c.Value())
	}
}

func TestCounter_Add(t *testing.T) {
	var c Counter
	c.Add(10)
	if c.Value() != 10 {
		t.Fatalf("expected 10, got %d", c.Value())
	}
}

func TestGauge_SetAndValue(t *testing.T) {
	var g Gauge
	g.Set(42)
	if g.Value() != 42 {
		t.Fatalf("expected 42, got %d", g.Value())
	}
	g.Set(-5)
	if g.Value() != -5 {
		t.Fatalf("expected -5, got %d", g.Value())
	}
}

func TestRegistry_CounterSameInstance(t *testing.T) {
	reg := New()
	c1 := reg.Counter("foo")
	c1.Inc()
	c2 := reg.Counter("foo")
	if c2.Value() != 1 {
		t.Fatalf("expected same instance, got value %d", c2.Value())
	}
}

func TestRegistry_GaugeSameInstance(t *testing.T) {
	reg := New()
	reg.Gauge("bar").Set(7)
	if reg.Gauge("bar").Value() != 7 {
		t.Fatal("expected same gauge instance")
	}
}

func TestRegistry_Snapshot(t *testing.T) {
	reg := New()
	reg.Counter("checks").Add(3)
	reg.Gauge("open_ports").Set(2)

	snap := reg.Snapshot()
	if snap["checks"] != 3 {
		t.Fatalf("expected checks=3, got %d", snap["checks"])
	}
	if snap["open_ports"] != 2 {
		t.Fatalf("expected open_ports=2, got %d", snap["open_ports"])
	}
}

func TestRegistry_ConcurrentSafety(t *testing.T) {
	reg := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reg.Counter("concurrent").Inc()
		}()
	}
	wg.Wait()
	if reg.Counter("concurrent").Value() != 100 {
		t.Fatalf("expected 100, got %d", reg.Counter("concurrent").Value())
	}
}
