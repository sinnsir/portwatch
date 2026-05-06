// Package dispatcher fans out port-state events to a set of registered
// handlers in a non-blocking, concurrent manner.
package dispatcher

import (
	"context"
	"sync"

	"github.com/user/portwatch/internal/monitor"
)

// Handler is any function that can receive a monitor.Event.
type Handler func(ctx context.Context, event monitor.Event)

// Dispatcher multiplexes a single event stream to many handlers.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers []Handler
}

// New returns a ready-to-use Dispatcher.
func New() *Dispatcher {
	return &Dispatcher{}
}

// Register appends h to the list of handlers that will be invoked for every
// event. It is safe to call Register after Dispatch has started.
func (d *Dispatcher) Register(h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = append(d.handlers, h)
}

// Dispatch reads events from ch until it is closed or ctx is cancelled,
// then calls every registered handler concurrently for each event.
// Dispatch blocks until the source channel is drained.
func (d *Dispatcher) Dispatch(ctx context.Context, ch <-chan monitor.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			d.fanOut(ctx, ev)
		}
	}
}

func (d *Dispatcher) fanOut(ctx context.Context, ev monitor.Event) {
	d.mu.RLock()
	hs := make([]Handler, len(d.handlers))
	copy(hs, d.handlers)
	d.mu.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(hs))
	for _, h := range hs {
		h := h
		go func() {
			defer wg.Done()
			h(ctx, ev)
		}()
	}
	wg.Wait()
}
