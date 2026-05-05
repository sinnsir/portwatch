// Package watchdog implements a lightweight self-healing watchdog for
// portwatch monitor goroutines.
//
// A Watchdog periodically calls a caller-supplied Probe function. When the
// probe returns false for Threshold consecutive intervals, the watchdog
// invokes a restart callback in a new goroutine so the caller can
// re-initialise the stalled component.
//
// Typical usage:
//
//	wd := watchdog.New(watchdog.DefaultPolicy(), func() bool {
//		return monitor.IsAlive()
//	}, func() {
//		monitor.Restart()
//	})
//	go wd.Run(ctx)
package watchdog
