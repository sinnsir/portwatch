// Package dispatcher provides a fan-out mechanism that delivers port-state
// events to multiple independent handlers concurrently.
//
// Usage:
//
//	d := dispatcher.New()
//	d.Register(func(ctx context.Context, ev monitor.Event) {
//		fmt.Println(ev)
//	})
//	d.Dispatch(ctx, eventChan)
//
// Every handler is invoked in its own goroutine; Dispatch blocks until all
// handlers for the current event have returned before pulling the next event
// from the channel.
package dispatcher
