package eventlog_test

import (
	"sync"
	"testing"

	"github.com/user/portwatch/internal/eventlog"
)

func TestAppend_ConcurrentSafety(t *testing.T) {
	l := eventlog.New(eventlog.Policy{MaxEntries: 100})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			l.Append(eventlog.Entry{Host: "localhost", Port: p, State: "open"})
		}(i)
	}
	wg.Wait()
	if l.Len() != 50 {
		t.Fatalf("expected 50 entries, got %d", l.Len())
	}
}

func TestSnapshot_ConcurrentWithAppend(t *testing.T) {
	l := eventlog.New(eventlog.Policy{MaxEntries: 200})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(p int) {
			defer wg.Done()
			l.Append(eventlog.Entry{Host: "localhost", Port: p, State: "open"})
		}(i)
		go func() {
			defer wg.Done()
			_ = l.Snapshot()
		}()
	}
	wg.Wait()
}

func TestQuery_ConcurrentSafety(t *testing.T) {
	l := eventlog.New(eventlog.Policy{MaxEntries: 100})
	for i := 0; i < 30; i++ {
		l.Append(eventlog.Entry{Host: "localhost", Port: i, State: "open"})
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = l.Query(eventlog.Filter{State: "open", Limit: 5})
		}()
	}
	wg.Wait()
}
