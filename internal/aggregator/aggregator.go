// Package aggregator batches port-state events over a time window and
// emits a single consolidated summary per flush cycle.
package aggregator

import (
	"sync"
	"time"
)

// Event represents a port state-change event to be aggregated.
type Event struct {
	Host  string
	Port  int
	State string
	At    time.Time
}

// Summary is the result of one flush cycle.
type Summary struct {
	Events    []Event
	FlushedAt time.Time
}

// Handler is called with a Summary on every flush.
type Handler func(Summary)

// Policy controls aggregator behaviour.
type Policy struct {
	// Window is the maximum time events are held before flushing.
	Window time.Duration
	// Capacity is the maximum number of events before an early flush.
	Capacity int
}

// DefaultPolicy returns sensible production defaults.
func DefaultPolicy() Policy {
	return Policy{
		Window:   5 * time.Second,
		Capacity: 100,
	}
}

// Aggregator collects events and flushes them periodically.
type Aggregator struct {
	policy  Policy
	handler Handler
	mu      sync.Mutex
	buf     []Event
	stop    chan struct{}
	wg      sync.WaitGroup
}

// New creates and starts a new Aggregator.
func New(p Policy, h Handler) *Aggregator {
	if p.Window <= 0 {
		p.Window = DefaultPolicy().Window
	}
	if p.Capacity <= 0 {
		p.Capacity = DefaultPolicy().Capacity
	}
	a := &Aggregator{
		policy:  p,
		handler: h,
		stop:    make(chan struct{}),
	}
	a.wg.Add(1)
	go a.loop()
	return a
}

// Add enqueues an event, triggering an early flush if capacity is reached.
func (a *Aggregator) Add(e Event) {
	if e.At.IsZero() {
		e.At = time.Now()
	}
	a.mu.Lock()
	a.buf = append(a.buf, e)
	flush := len(a.buf) >= a.policy.Capacity
	a.mu.Unlock()
	if flush {
		a.flush()
	}
}

// Stop flushes any remaining events and shuts down the background goroutine.
func (a *Aggregator) Stop() {
	close(a.stop)
	a.wg.Wait()
	a.flush()
}

func (a *Aggregator) loop() {
	defer a.wg.Done()
	ticker := time.NewTicker(a.policy.Window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.flush()
		case <-a.stop:
			return
		}
	}
}

func (a *Aggregator) flush() {
	a.mu.Lock()
	if len(a.buf) == 0 {
		a.mu.Unlock()
		return
	}
	events := make([]Event, len(a.buf))
	copy(events, a.buf)
	a.buf = a.buf[:0]
	a.mu.Unlock()
	a.handler(Summary{Events: events, FlushedAt: time.Now()})
}
