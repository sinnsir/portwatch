// Package schedule provides a lightweight periodic scheduler for portwatch.
//
// A Scheduler fires a user-supplied job function on a configurable interval.
// The SkipIfBusy policy option prevents overlapping executions so that a
// slow job does not queue up behind itself.
//
// Basic usage:
//
//	sched := schedule.New(schedule.DefaultPolicy(), func(ctx context.Context) {
//		// perform periodic work
//	})
//	go sched.Run(ctx)
package schedule
