package schedule_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/schedule"
)

func TestScheduler_ConcurrentSafety(t *testing.T) {
	var count int64
	policy := schedule.Policy{Interval: 5 * time.Millisecond, SkipIfBusy: false}
	sched := schedule.New(policy, func(_ context.Context) {
		atomic.AddInt64(&count, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	sched.Run(ctx)

	if atomic.LoadInt64(&count) < 5 {
		t.Fatalf("expected several ticks under concurrent load, got %d", count)
	}
}

func TestScheduler_JobReceivesContext(t *testing.T) {
	results := make(chan error, 4)
	policy := schedule.Policy{Interval: 15 * time.Millisecond, SkipIfBusy: true}
	sched := schedule.New(policy, func(ctx context.Context) {
		results <- ctx.Err()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
	defer cancel()
	sched.Run(ctx)
	close(results)

	for err := range results {
		if err != nil {
			t.Fatalf("job received cancelled context mid-tick: %v", err)
		}
	}
}
