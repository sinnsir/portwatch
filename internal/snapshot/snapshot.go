// Package snapshot provides periodic state snapshots of monitored ports,
// persisting the latest known state to disk for recovery after restart.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry holds the last known state for a single port.
type Entry struct {
	Port      int       `json:"port"`
	Host      string    `json:"host"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Snapshot manages in-memory port state and its persistence.
type Snapshot struct {
	mu      sync.RWMutex
	entries map[string]Entry // key: "host:port"
	path    string
}

// New creates a Snapshot that persists state to the given file path.
// Existing data is loaded on creation if the file exists.
func New(path string) (*Snapshot, error) {
	s := &Snapshot{
		entries: make(map[string]Entry),
		path:    path,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Set records or updates the state for a host/port pair and flushes to disk.
func (s *Snapshot) Set(host string, port int, state string) error {
	key := entryKey(host, port)
	s.mu.Lock()
	s.entries[key] = Entry{
		Port:      port,
		Host:      host,
		State:     state,
		UpdatedAt: time.Now().UTC(),
	}
	s.mu.Unlock()
	return s.flush()
}

// Get returns the last known Entry for a host/port pair and whether it exists.
func (s *Snapshot) Get(host string, port int) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[entryKey(host, port)]
	return e, ok
}

// All returns a copy of all stored entries.
func (s *Snapshot) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

func (s *Snapshot) flush() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.entries, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Snapshot) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Unmarshal(data, &s.entries)
}

func entryKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
