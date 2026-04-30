// Package ratelimit provides a token-bucket rate limiter for controlling
// how frequently port-check actions (webhooks, commands) may fire.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a per-key token-bucket rate limiter.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // max events per window
	window   time.Duration // rolling window size
}

type bucket struct {
	tokens    int
	windowEnd time.Time
}

// New creates a Limiter that allows up to rate events per window duration per key.
func New(rate int, window time.Duration) *Limiter {
	if rate <= 0 {
		rate = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

// Allow reports whether the event identified by key is permitted under the
// current rate limit. It returns false when the bucket is exhausted.
func (l *Limiter) Allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.After(b.windowEnd) {
		l.buckets[key] = &bucket{
			tokens:    l.rate - 1,
			windowEnd: now.Add(l.window),
		}
		return true
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

// Reset clears the bucket for the given key, allowing it to fire immediately.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// Remaining returns how many tokens are left in the current window for key.
func (l *Limiter) Remaining(key string) int {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok || now.After(b.windowEnd) {
		return l.rate
	}
	return b.tokens
}
