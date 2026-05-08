// Package resolver maps port-check events to named services using a
// configurable host:port → service-name registry. It enriches events with a
// human-readable service label so downstream handlers (notifier, history, API)
// can display meaningful names instead of raw addresses.
package resolver

import (
	"fmt"
	"sync"
)

// Entry associates a host:port pair with a service name.
type Entry struct {
	Host    string
	Port    int
	Service string
}

// Resolver holds the service registry and provides thread-safe lookups.
type Resolver struct {
	mu      sync.RWMutex
	entries map[string]string // key: "host:port"
}

// New creates a Resolver pre-populated with the supplied entries.
// Duplicate keys are silently overwritten by the last entry in the slice.
func New(entries []Entry) *Resolver {
	r := &Resolver{
		entries: make(map[string]string, len(entries)),
	}
	for _, e := range entries {
		r.entries[key(e.Host, e.Port)] = e.Service
	}
	return r
}

// Lookup returns the service name registered for host:port.
// The second return value is false when no entry exists.
func (r *Resolver) Lookup(host string, port int) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	svc, ok := r.entries[key(host, port)]
	return svc, ok
}

// Register adds or replaces a service mapping at runtime.
func (r *Resolver) Register(host string, port int, service string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[key(host, port)] = service
}

// Remove deletes the mapping for host:port. It is a no-op if the key is
// not present.
func (r *Resolver) Remove(host string, port int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, key(host, port))
}

// Len returns the number of registered entries.
func (r *Resolver) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

func key(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
