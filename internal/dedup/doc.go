// Package dedup implements TTL-based event deduplication for portwatch.
//
// A Deduplicator tracks keys (typically composed from host + port + state) and
// suppresses identical entries seen within a configurable time window.  This
// prevents webhook floods when a port flaps rapidly or when the monitor polls
// faster than the downstream consumer can process.
//
// Usage:
//
//	d := dedup.New(dedup.DefaultPolicy)
//	if !d.IsDuplicate(key) {
//		// forward event
//	}
package dedup
