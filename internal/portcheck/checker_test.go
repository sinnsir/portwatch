package portcheck_test

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/portcheck"
)

// startTCPServer opens a random TCP listener and returns its port + a close func.
func startTCPServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	port, _ := strconv.Atoi(ln.Addr().(*net.TCPAddr).Port.String())
	// Accept connections in background so dial doesn't block.
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()
	return port, func() { _ = ln.Close() }
}

func TestChecker_OpenPort(t *testing.T) {
	port, stop := startTCPServer(t)
	defer stop()

	c := portcheck.NewChecker(2 * time.Second)
	res := c.Check("127.0.0.1", port)

	if res.State != portcheck.StateOpen {
		t.Errorf("expected StateOpen, got %s (err: %v)", res.State, res.Err)
	}
	if res.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestChecker_ClosedPort(t *testing.T) {
	// Port 1 is almost certainly closed and unprivileged access denied.
	c := portcheck.NewChecker(500 * time.Millisecond)
	res := c.Check("127.0.0.1", 1)

	if res.State != portcheck.StateClosed {
		t.Errorf("expected StateClosed, got %s", res.State)
	}
	if res.Err == nil {
		t.Error("expected non-nil error for closed port")
	}
}

func TestStateString(t *testing.T) {
	cases := []struct {
		s    portcheck.State
		want string
	}{
		{portcheck.StateOpen, "open"},
		{portcheck.StateClosed, "closed"},
		{portcheck.StateUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
