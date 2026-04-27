// Package history maintains a ring-buffer of recent port state-change events
// so that the CLI can surface a brief audit trail without an external database.
package history

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/portcheck"
)

// Event records a single state transition for a monitored port.
type Event struct {
	Port      int
	Host      string
	OldState  portcheck.State
	NewState  portcheck.State
	Timestamp time.Time
}

// History is a thread-safe, fixed-capacity ring buffer of Events.
type History struct {
	mu       sync.Mutex
	buf      []Event
	cap      int
	head     int
	count    int
}

// New creates a History that retains at most capacity events.
// If capacity is <= 0 it defaults to 100.
func New(capacity int) *History {
	if capacity <= 0 {
		capacity = 100
	}
	return &History{
		buf: make([]Event, capacity),
		cap: capacity,
	}
}

// Record appends an event to the ring buffer, overwriting the oldest entry
// when the buffer is full.
func (h *History) Record(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	index := (h.head + h.count) % h.cap
	h.buf[index] = e
	if h.count < h.cap {
		h.count++
	} else {
		// Buffer full — advance head to discard oldest.
		h.head = (h.head + 1) % h.cap
	}
}

// Snapshot returns a copy of all stored events in chronological order
// (oldest first).
func (h *History) Snapshot() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]Event, h.count)
	for i := 0; i < h.count; i++ {
		out[i] = h.buf[(h.head+i)%h.cap]
	}
	return out
}

// Len returns the number of events currently stored.
func (h *History) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}
