package filter_test

import (
	"sync"
	"testing"

	"github.com/patrickward/portwatch/internal/filter"
)

// TestAllow_ConcurrentSafety verifies that Allow is safe to call from multiple
// goroutines simultaneously (the filter is read-only after construction).
func TestAllow_ConcurrentSafety(t *testing.T) {
	f := filter.New(
		filter.WithIncludeTags("prod"),
		filter.WithExcludeTags("disabled"),
	)

	e := entry("svc", "prod")
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if !f.Allow(e) {
				t.Errorf("expected entry to be allowed")
			}
		}()
	}
	wg.Wait()
}

// TestAllow_MultipleEntriesMixedOutcomes ensures a single Filter correctly
// differentiates between a batch of entries with varying tags.
func TestAllow_MultipleEntriesMixedOutcomes(t *testing.T) {
	f := filter.New(
		filter.WithIncludeTags("prod"),
		filter.WithExcludeTags("disabled"),
	)

	cases := []struct {
		e    filter.Entry
		want bool
	}{
		{entry("a", "prod"), true},
		{entry("b", "staging"), false},
		{entry("c", "prod", "disabled"), false},
		{entry("d", "prod", "beta"), true},
		{entry("e"), false},
	}

	for _, tc := range cases {
		got := f.Allow(tc.e)
		if got != tc.want {
			t.Errorf("Allow(%q tags=%v) = %v, want %v",
				tc.e.Label, tc.e.Tags, got, tc.want)
		}
	}
}
