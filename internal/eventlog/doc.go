// Package eventlog provides a bounded, thread-safe in-memory log of port
// state-change events for portwatch.
//
// # Overview
//
// An [Log] retains up to Policy.MaxEntries events (default 500). When the
// capacity is reached the oldest entry is silently evicted, giving a sliding
// window of recent activity.
//
// # Usage
//
//	log := eventlog.New(eventlog.Policy{MaxEntries: 200})
//	log.Append(eventlog.Entry{Host: "localhost", Port: 8080, State: "open"})
//	entries := log.Snapshot()
package eventlog
