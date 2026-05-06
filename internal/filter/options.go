package filter

import "strings"

// Option configures a Filter.
type Option func(*Filter)

// WithIncludeTags restricts matching to entries that carry at least one of the
// provided tags (case-insensitive).
func WithIncludeTags(tags ...string) Option {
	return func(f *Filter) {
		for _, t := range tags {
			f.includeTags[strings.ToLower(t)] = struct{}{}
		}
	}
}

// WithExcludeTags rejects entries that carry any of the provided tags
// (case-insensitive).
func WithExcludeTags(tags ...string) Option {
	return func(f *Filter) {
		for _, t := range tags {
			f.excludeTags[strings.ToLower(t)] = struct{}{}
		}
	}
}

// WithPattern sets a glob pattern (path.Match syntax) matched against the
// entry Label, falling back to Host when Label is empty.
func WithPattern(pattern string) Option {
	return func(f *Filter) {
		f.pattern = pattern
	}
}
