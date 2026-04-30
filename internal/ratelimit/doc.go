// Package ratelimit implements a thread-safe, per-key token-bucket rate limiter
// used to throttle webhook and shell-command actions triggered by port state
// changes in portwatch.
//
// Usage:
//
//	// Allow up to 5 events per port per minute.
//	 limiter := ratelimit.New(5, time.Minute)
//
//	 if limiter.Allow("port:8080:open") {
//	     // fire webhook / run command
//	 }
//
// Each unique key maintains its own independent token bucket. Buckets are
// reset automatically when their window expires, requiring no background
// goroutine.
package ratelimit
