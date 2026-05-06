package dispatcher_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/dispatcher"
	"github.com/user/portwatch/internal/monitor"
	"github.com/user/portwatch/internal/portcheck"
)

func TestDispatch_ConcurrentSafety(t *testing.T) {
	d := dispatcher.New()
	var total atomic.Int64

	const handlers = 20
	const events = 50

	for i := 0; i < handlers; i++ {
		d.Register(func(_ context.Context, _ monitor.Event) {
			total.Add(1)
		})
	}

	ch := make(chan monitor.Event, events)
	for i := 0; i < events; i++ {
		ch <- monitor.Event{
			Host:  "127.0.0.1",
			Port:  8000 + i,
			State: portcheck.StateOpen,
		}
	}
	close(ch)

	d.Dispatch(context.Background(), ch)

	want := int64(handlers * events)
	if total.Load() != want {
		t.Fatalf("expected %d total handler invocations, got %d", want, total.Load())
	}
}

func TestDispatch_HandlerReceivesCorrectEvent(t *testing.T) {
	d := dispatcher.New()
	received := make(chan monitor.Event, 1)

	d.Register(func(_ context.Context, ev monitor.Event) {
		received <- ev
	})

	ch := make(chan monitor.Event, 1)
	want := monitor.Event{
		Host:  "example.com",
		Port:  443,
		State: portcheck.StateClosed,
	}
	ch <- want
	close(ch)

	d.Dispatch(context.Background(), ch)

	select {
	case got := <-received:
		if got != want {
			t.Fatalf("event mismatch: got %+v, want %+v", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("handler was never invoked")
	}
}
