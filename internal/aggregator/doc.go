// Package aggregator provides a time-windowed event aggregator for port-state
// changes. Events are collected in an internal buffer and flushed either when
// the configured window elapses or when the buffer reaches its capacity,
// whichever comes first.
//
// Typical usage:
//
//	a := aggregator.New(aggregator.DefaultPolicy(), func(s aggregator.Summary) {
//		for _, e := range s.Events {
//			log.Printf("port %d on %s is %s", e.Port, e.Host, e.State)
//		}
//	})
//	defer a.Stop()
//
//	a.Add(aggregator.Event{Host: "localhost", Port: 8080, State: "open"})
package aggregator
