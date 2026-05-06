// Package filter provides lightweight tag-based and glob-pattern filtering
// for port-watch entries.
//
// A Filter is constructed once with a set of Option values and then used to
// decide whether a given Entry should be monitored:
//
//	f := filter.New(
//		filter.WithIncludeTags("prod"),
//		filter.WithExcludeTags("disabled"),
//		filter.WithPattern("api-*"),
//	)
//	if f.Allow(entry) {
//		// start monitoring
//	}
//
// All rules are AND-combined: an entry must satisfy every configured
// constraint to be allowed through.
package filter
