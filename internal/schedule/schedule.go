// Package schedule provides a cron-style periodic scheduler that fires
// callbacks at a fixed interval and supports graceful shutdown.
package schedule

import (
	"context"
	"sync"
	"time"
)

// Policy controls scheduler behaviour.
type Policy struct {
	// Interval between successive ticks.
	Interval time.Duration
	// SkipIfBusy prevents overlapping executions when true.
	SkipIfBusy bool
}

// DefaultPolicy returns a sensible default policy.
func DefaultPolicy() Policy {
	return Policy{
		Interval:   30 * time.Second,
		SkipIfBusy: true,
	}
}

// Scheduler fires a Job function on a fixed interval until stopped.
type Scheduler struct {
	policy Policy
	job    func(ctx context.Context)
	mu     sync.Mutex
	busy   bool
}

// New creates a Scheduler with the given policy and job.
// If policy.Interval is zero or negative the default interval is used.
func New(policy Policy, job func(ctx context.Context)) *Scheduler {
	if policy.Interval <= 0 {
		policy.Interval = DefaultPolicy().Interval
	}
	return &Scheduler{policy: policy, job: job}
}

// Run starts the scheduler and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.policy.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	if s.policy.SkipIfBusy {
		s.mu.Lock()
		if s.busy {
			s.mu.Unlock()
			return
		}
		s.busy = true
		s.mu.Unlock()

		defer func() {
			s.mu.Lock()
			s.busy = false
			s.mu.Unlock()
		}()
	}
	s.job(ctx)
}
