// Package cache provides a lightweight TTL-based in-memory key/value store
// used to avoid redundant work across the portwatch pipeline.
package cache

import (
	"sync"
	"time"
)

// entry holds a cached value and its expiry deadline.
type entry[V any] struct {
	value   V
	expiry  time.Time
}

// Cache is a generic TTL cache safe for concurrent use.
type Cache[K comparable, V any] struct {
	mu  sync.Mutex
	items map[K]entry[V]
	ttl  time.Duration
}

// New returns a Cache with the given TTL. A zero or negative TTL defaults to
// 30 seconds.
func New[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &Cache[K, V]{
		items: make(map[K]entry[V]),
		ttl:   ttl,
	}
}

// Set stores value under key, resetting its TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry[V]{value: value, expiry: time.Now().Add(c.ttl)}
}

// Get returns the value for key and whether it was found and still valid.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiry) {
		delete(c.items, key)
		var zero V
		return zero, false
	}
	return e.value, true
}

// Delete removes key from the cache immediately.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Purge removes all expired entries and returns the number of entries evicted.
func (c *Cache[K, V]) Purge() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	evicted := 0
	for k, e := range c.items {
		if now.After(e.expiry) {
			delete(c.items, k)
			evicted++
		}
	}
	return evicted
}

// Len returns the number of entries currently in the cache (including expired
// ones not yet purged).
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}
