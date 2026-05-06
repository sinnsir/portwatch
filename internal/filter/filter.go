// Package filter provides tag-based and pattern-based filtering for port entries.
// It allows selectively enabling or disabling monitoring targets at runtime
// based on user-defined labels attached to each port configuration.
package filter

import (
	"path"
	"strings"
)

// Entry represents a monitorable target with optional tags and a host:port label.
type Entry struct {
	Host string
	Port int
	Tags []string
	Label string // optional human-readable name
}

// Filter holds compiled inclusion/exclusion rules.
type Filter struct {
	includeTags map[string]struct{}
	excludeTags map[string]struct{}
	pattern     string // glob pattern matched against Label or "host:port"
}

// New creates a Filter from the given options.
func New(opts ...Option) *Filter {
	f := &Filter{
		includeTags: make(map[string]struct{}),
		excludeTags: make(map[string]struct{}),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Allow returns true when the entry passes all configured filter rules.
// An entry is allowed when:
//  1. It matches the glob pattern (if one is set).
//  2. It carries at least one included tag (if include tags are set).
//  3. It carries none of the excluded tags.
func (f *Filter) Allow(e Entry) bool {
	if f.pattern != "" {
		target := e.Label
		if target == "" {
			target = e.Host
		}
		matched, err := path.Match(f.pattern, target)
		if err != nil || !matched {
			return false
		}
	}

	if len(f.includeTags) > 0 {
		if !hasAny(e.Tags, f.includeTags) {
			return false
		}
	}

	if len(f.excludeTags) > 0 {
		if hasAny(e.Tags, f.excludeTags) {
			return false
		}
	}

	return true
}

func hasAny(tags []string, set map[string]struct{}) bool {
	for _, t := range tags {
		if _, ok := set[strings.ToLower(t)]; ok {
			return true
		}
	}
	return false
}
