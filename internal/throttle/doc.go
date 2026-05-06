// Package throttle implements a shared token-bucket throughput limiter for
// portwatch action dispatch.
//
// # Overview
//
// When many ports change state simultaneously (e.g. after a network blip)
// portwatch could fire dozens of webhooks in a single instant. The throttle
// package provides a single shared bucket that caps the global dispatch rate,
// protecting downstream receivers.
//
// # Usage
//
//	t := throttle.New(throttle.DefaultPolicy())
//	if t.Allow() {
//		// dispatch action
//	}
package throttle
