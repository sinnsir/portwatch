// Package dedup provides event deduplication based on a configurable TTL window.
// Identical events (same port, host, and state) seen within the window are
// suppressed so that downstream handlers are not flooded with repeated signals.
package dedup

import (
	"sync"
	"time"
)

// Policy controls deduplication behaviour.
type Policy struct {
	// Window is the duration during which a duplicate key is suppressed.
	// Zero or negative values default to DefaultPolicy.Window.
	Window time.Duration
}

// DefaultPolicy is the out-of-the-box deduplication policy.
var DefaultPolicy = Policy{
	Window: 30 * time.Second,
}

type entry struct {
	expiresAt time.Time
}

// Deduplicator suppresses repeated events within a TTL window.
type Deduplicator struct {
	mu      sync.Mutex
	window  time.Duration
	seen    map[string]entry
	nowFunc func() time.Time
}

// New returns a Deduplicator configured with p.
// If p.Window <= 0 the default window is used.
func New(p Policy) *Deduplicator {
	w := p.Window
	if w <= 0 {
		w = DefaultPolicy.Window
	}
	return &Deduplicator{
		window:  w,
		seen:    make(map[string]entry),
		nowFunc: time.Now,
	}
}

// IsDuplicate returns true when key was already seen within the window.
// The first call for a new (or expired) key returns false and records it.
func (d *Deduplicator) IsDuplicate(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.nowFunc()
	if e, ok := d.seen[key]; ok && now.Before(e.expiresAt) {
		return true
	}
	d.seen[key] = entry{expiresAt: now.Add(d.window)}
	return false
}

// Purge removes all expired entries. Call periodically to prevent unbounded
// memory growth in long-running daemons.
func (d *Deduplicator) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.nowFunc()
	for k, e := range d.seen {
		if now.After(e.expiresAt) {
			delete(d.seen, k)
		}
	}
}

// Len returns the number of currently tracked keys (including expired ones
// not yet purged).
func (d *Deduplicator) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}
