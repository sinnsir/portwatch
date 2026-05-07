// Package fanout provides a generic, concurrency-safe broadcast primitive.
//
// A Fanout[T] holds an ordered list of handlers. When Publish is called with
// an event, every handler is invoked concurrently in its own goroutine.
// Publish blocks until all handlers return and collects any non-nil errors
// into a slice that is returned to the caller.
//
// Typical usage:
//
//	f := fanout.New[portcheck.Event]()
//	f.Register(notifySlack)
//	f.Register(recordHistory)
//
//	errs := f.Publish(ctx, event)
//	for _, err := range errs {
//		log.Println("handler error:", err)
//	}
//
// Handlers are safe to register after construction; Register and Publish
// may be called from multiple goroutines simultaneously.
package fanout
