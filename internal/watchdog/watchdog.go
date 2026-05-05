// Package watchdog provides a self-healing watchdog that restarts
// stalled monitor goroutines when they stop reporting heartbeats.
package watchdog

import (
	"context"
	"log"
	"sync"
	"time"
)

// Probe is a function that returns true if the monitored component is alive.
type Probe func() bool

// Policy controls watchdog behaviour.
type Policy struct {
	// Interval is how often the watchdog checks liveness.
	Interval time.Duration
	// Threshold is the number of consecutive failed probes before action.
	Threshold int
}

// DefaultPolicy returns a sensible production policy.
func DefaultPolicy() Policy {
	return Policy{
		Interval:  10 * time.Second,
		Threshold: 3,
	}
}

// Watchdog periodically probes a component and calls a restart function
// when the component appears unhealthy.
type Watchdog struct {
	policy  Policy
	probe   Probe
	restart func()

	mu       sync.Mutex
	failures int
}

// New creates a Watchdog with the given policy, liveness probe, and restart hook.
func New(p Policy, probe Probe, restart func()) *Watchdog {
	if p.Interval <= 0 {
		p.Interval = DefaultPolicy().Interval
	}
	if p.Threshold <= 0 {
		p.Threshold = DefaultPolicy().Threshold
	}
	return &Watchdog{
		policy:  p,
		probe:   probe,
		restart: restart,
	}
}

// Run starts the watchdog loop. It blocks until ctx is cancelled.
func (w *Watchdog) Run(ctx context.Context) {
	ticker := time.NewTicker(w.policy.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick()
		}
	}
}

func (w *Watchdog) tick() {
	alive := w.probe()

	w.mu.Lock()
	defer w.mu.Unlock()

	if alive {
		w.failures = 0
		return
	}

	w.failures++
	log.Printf("watchdog: probe failed (%d/%d)", w.failures, w.policy.Threshold)

	if w.failures >= w.policy.Threshold {
		log.Printf("watchdog: threshold reached, triggering restart")
		w.failures = 0
		go w.restart()
	}
}

// Failures returns the current consecutive failure count.
func (w *Watchdog) Failures() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.failures
}
