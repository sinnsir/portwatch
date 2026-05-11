// Package eventlog provides a structured, append-only log of port state-change
// events with configurable maximum capacity and optional persistence via a
// pluggable writer.
package eventlog

import (
	"sync"
	"time"
)

// Entry represents a single logged port event.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	State     string    `json:"state"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Policy controls the behaviour of the event log.
type Policy struct {
	// MaxEntries is the maximum number of entries retained in memory.
	// Zero or negative values default to DefaultMaxEntries.
	MaxEntries int
}

// DefaultMaxEntries is used when Policy.MaxEntries is not positive.
const DefaultMaxEntries = 500

// Log is a thread-safe, bounded in-memory event log.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	max     int
}

// New returns a new Log configured by p.
func New(p Policy) *Log {
	max := p.MaxEntries
	if max <= 0 {
		max = DefaultMaxEntries
	}
	return &Log{
		entries: make([]Entry, 0, min(max, 64)),
		max:     max,
	}
}

// Append adds e to the log. If the log is at capacity the oldest entry is
// evicted.
func (l *Log) Append(e Entry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= l.max {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, e)
}

// Snapshot returns a shallow copy of all current entries, oldest first.
func (l *Log) Snapshot() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Len returns the number of entries currently stored.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}

// Clear removes all entries from the log.
func (l *Log) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = l.entries[:0]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
