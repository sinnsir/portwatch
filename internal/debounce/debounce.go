// Package debounce provides a state-change debouncer that suppresses
// transient port flaps by requiring a state to be observed consecutively
// before forwarding it to downstream consumers.
package debounce

import (
	"sync"
	"time"
)

// Policy controls debounce behaviour.
type Policy struct {
	// Threshold is the number of consecutive identical observations required
	// before the state is considered stable and emitted.
	Threshold int

	// Window is the maximum duration within which consecutive observations
	// must occur. Observations older than Window reset the counter.
	Window time.Duration
}

// DefaultPolicy returns sensible defaults: 3 consecutive hits within 5 s.
func DefaultPolicy() Policy {
	return Policy{
		Threshold: 3,
		Window:    5 * time.Second,
	}
}

type entry struct {
	value     bool
	count     int
	lastSeen  time.Time
}

// Debouncer tracks per-key state observations and reports whether a new
// observation should be forwarded downstream.
type Debouncer struct {
	policy Policy
	mu     sync.Mutex
	state  map[string]*entry
	now    func() time.Time
}

// New creates a Debouncer with the given policy.
// If policy.Threshold < 1 it is clamped to 1.
// If policy.Window <= 0 it is set to 5 s.
func New(p Policy) *Debouncer {
	if p.Threshold < 1 {
		p.Threshold = 1
	}
	if p.Window <= 0 {
		p.Window = 5 * time.Second
	}
	return &Debouncer{
		policy: p,
		state:  make(map[string]*entry),
		now:    time.Now,
	}
}

// Observe records an observation of value for key and returns true when the
// observation should be forwarded (i.e. threshold consecutive identical
// values within the window have been seen).
func (d *Debouncer) Observe(key string, value bool) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	e, ok := d.state[key]
	if !ok {
		e = &entry{}
		d.state[key] = e
	}

	// Reset counter if value changed or window expired.
	if e.value != value || now.Sub(e.lastSeen) > d.policy.Window {
		e.value = value
		e.count = 0
	}

	e.count++
	e.lastSeen = now

	return e.count >= d.policy.Threshold
}

// Reset removes all tracked state, useful when a monitor is stopped.
func (d *Debouncer) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = make(map[string]*entry)
}
