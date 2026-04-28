// Package history provides an in-memory ring buffer that records port state
// change events. It is safe for concurrent use.
//
// # Recording events
//
// Create a History with New, then call Record each time a port state changes.
// The buffer overwrites the oldest entry once capacity is reached.
//
// # Exporting
//
// Snapshots can be serialised via ExportJSON (machine-readable) or
// ExportTable (human-readable CLI output).
package history
