package monitor

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// newTestMonitor creates a Monitor configured with a single port entry for use
// in tests. The returned key is the address string used in m.states.
func newTestMonitor(t *testing.T, port int) (m *Monitor, key string) {
	t.Helper()
	cfg := &config.Config{
		Interval: 1,
		Ports: []config.PortEntry{
			{Host: "127.0.0.1", Port: port},
		},
	}
	m = New(cfg)
	key = "127.0.0.1:" + strconv.Itoa(port)
	return m, key
}

func TestMonitor_DetectsInitialState(t *testing.T) {
	port := freePort(t)
	m, key := newTestMonitor(t, port)

	// checkAll should populate initial state without panicking.
	m.checkAll()

	if _, ok := m.states[key]; !ok {
		t.Errorf("expected state to be recorded for %s", key)
	}
}

func TestMonitor_DetectsStateChange(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port

	m, key := newTestMonitor(t, port)
	m.checkAll() // initial: open

	initialState := m.states[key]
	if initialState.String() != "open" {
		t.Fatalf("expected open, got %s", initialState)
	}

	l.Close()
	time.Sleep(50 * time.Millisecond)

	m.checkAll() // should now be closed
	newState := m.states[key]
	if newState.String() != "closed" {
		t.Fatalf("expected closed after listener closed, got %s", newState)
	}
}

func TestMonitor_StopUnblocks(t *testing.T) {
	cfg := &config.Config{
		Interval: 60,
		Ports:    []config.PortEntry{},
	}
	m := New(cfg)

	done := make(chan struct{})
	go func() {
		m.Start()
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	m.Stop()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not stop in time")
	}
}
