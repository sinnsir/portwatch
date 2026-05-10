// Package ticker provides a context-aware interval ticker with optional jitter.
//
// It is used by the monitor and schedule subsystems to drive periodic port
// checks without coupling them to time.Ticker directly. Jitter spreads
// concurrent checks across a small time window to avoid thundering-herd
// effects when many ports are monitored simultaneously.
//
// Basic usage:
//
//	tk := ticker.New(ticker.Policy{
//		Interval: 5 * time.Second,
//		Jitter:   500 * time.Millisecond,
//	})
//	ch := make(chan time.Time, 1)
//	go tk.Run(ctx, ch)
//	for t := range ch { ... }
package ticker
