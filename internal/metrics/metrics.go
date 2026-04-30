// Package metrics provides lightweight in-process counters and gauges
// for tracking portwatch runtime statistics such as check counts,
// state changes, webhook deliveries, and command executions.
package metrics

import (
	"sync"
	"sync/atomic"
)

// Counter is a monotonically increasing uint64 counter.
type Counter struct{ v uint64 }

// Inc increments the counter by 1.
func (c *Counter) Inc() { atomic.AddUint64(&c.v, 1) }

// Add increments the counter by n.
func (c *Counter) Add(n uint64) { atomic.AddUint64(&c.v, n) }

// Value returns the current counter value.
func (c *Counter) Value() uint64 { return atomic.LoadUint64(&c.v) }

// Gauge holds a int64 value that can go up or down.
type Gauge struct {
	mu sync.Mutex
	v  int64
}

// Set sets the gauge to v.
func (g *Gauge) Set(v int64) {
	g.mu.Lock()
	g.v = v
	g.mu.Unlock()
}

// Value returns the current gauge value.
func (g *Gauge) Value() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.v
}

// Registry holds all named metrics for the process.
type Registry struct {
	mu       sync.RWMutex
	counters map[string]*Counter
	gauges   map[string]*Gauge
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
	}
}

// Counter returns (creating if necessary) the named counter.
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{}
	r.counters[name] = c
	return c
}

// Gauge returns (creating if necessary) the named gauge.
func (r *Registry) Gauge(name string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.gauges[name]; ok {
		return g
	}
	g := &Gauge{}
	r.gauges[name] = g
	return g
}

// Snapshot returns a point-in-time copy of all metric values.
func (r *Registry) Snapshot() map[string]int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]int64, len(r.counters)+len(r.gauges))
	for k, c := range r.counters {
		out[k] = int64(c.Value())
	}
	for k, g := range r.gauges {
		out[k] = g.Value()
	}
	return out
}
