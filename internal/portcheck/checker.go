package portcheck

import (
	"fmt"
	"net"
	"time"
)

// State represents the availability state of a port.
type State int

const (
	StateUnknown State = iota
	StateOpen
	StateClosed
)

func (s State) String() string {
	switch s {
	case StateOpen:
		return "open"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Result holds the outcome of a single port check.
type Result struct {
	Host    string
	Port    int
	State   State
	Latency time.Duration
	Err     error
}

// Checker probes TCP port availability.
type Checker struct {
	Timeout time.Duration
}

// NewChecker creates a Checker with the given dial timeout.
func NewChecker(timeout time.Duration) *Checker {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Checker{Timeout: timeout}
}

// Check dials host:port and returns a Result.
func (c *Checker) Check(host string, port int) Result {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, c.Timeout)
	latency := time.Since(start)

	if err != nil {
		return Result{
			Host:    host,
			Port:    port,
			State:   StateClosed,
			Latency: latency,
			Err:     err,
		}
	}

	_ = conn.Close()
	return Result{
		Host:    host,
		Port:    port,
		State:   StateOpen,
		Latency: latency,
	}
}
