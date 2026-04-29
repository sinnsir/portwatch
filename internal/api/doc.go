// Package api implements a lightweight read-only HTTP API for portwatch.
//
// It exposes two endpoints:
//
//	GET /healthz              — liveness probe, returns {"status":"ok"}
//	GET /api/v1/events        — JSON array of recorded port-state events
//	                            accepts an optional ?limit=N query parameter
//
// The server is intentionally read-only; all state mutations happen through
// the monitor and notifier pipelines.
package api
