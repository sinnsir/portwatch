// Package metrics exposes lightweight in-process counters and gauges used
// throughout portwatch to track operational statistics.
//
// A Registry is the central collection point. Callers obtain named Counter
// or Gauge instances from it; all operations are safe for concurrent use.
//
// Typical usage:
//
//	reg := metrics.New()
//	checks := reg.Counter("port_checks_total")
//	checks.Inc()
//
// Snapshots can be serialised and served via the HTTP API or written to a
// log for external scraping.
package metrics
