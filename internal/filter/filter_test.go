package filter_test

import (
	"testing"

	"github.com/patrickward/portwatch/internal/filter"
)

func entry(label string, tags ...string) filter.Entry {
	return filter.Entry{Host: "localhost", Port: 8080, Label: label, Tags: tags}
}

func TestAllow_NoRules_AcceptsAll(t *testing.T) {
	f := filter.New()
	if !f.Allow(entry("anything")) {
		t.Fatal("expected entry to be allowed when no rules are set")
	}
}

func TestAllow_IncludeTag_Match(t *testing.T) {
	f := filter.New(filter.WithIncludeTags("prod"))
	if !f.Allow(entry("svc", "prod", "api")) {
		t.Fatal("expected entry with matching include tag to be allowed")
	}
}

func TestAllow_IncludeTag_NoMatch(t *testing.T) {
	f := filter.New(filter.WithIncludeTags("prod"))
	if f.Allow(entry("svc", "staging")) {
		t.Fatal("expected entry without include tag to be rejected")
	}
}

func TestAllow_ExcludeTag_Rejects(t *testing.T) {
	f := filter.New(filter.WithExcludeTags("disabled"))
	if f.Allow(entry("svc", "prod", "disabled")) {
		t.Fatal("expected entry with excluded tag to be rejected")
	}
}

func TestAllow_ExcludeTag_Absent_Passes(t *testing.T) {
	f := filter.New(filter.WithExcludeTags("disabled"))
	if !f.Allow(entry("svc", "prod")) {
		t.Fatal("expected entry without excluded tag to be allowed")
	}
}

func TestAllow_Pattern_Match(t *testing.T) {
	f := filter.New(filter.WithPattern("api-*"))
	if !f.Allow(entry("api-gateway")) {
		t.Fatal("expected label matching glob to be allowed")
	}
}

func TestAllow_Pattern_NoMatch(t *testing.T) {
	f := filter.New(filter.WithPattern("api-*"))
	if f.Allow(entry("worker-queue")) {
		t.Fatal("expected label not matching glob to be rejected")
	}
}

func TestAllow_Pattern_FallsBackToHost(t *testing.T) {
	f := filter.New(filter.WithPattern("local*"))
	e := filter.Entry{Host: "localhost", Port: 9090}
	if !f.Allow(e) {
		t.Fatal("expected host to be used when label is empty")
	}
}

func TestAllow_CaseInsensitiveTags(t *testing.T) {
	f := filter.New(filter.WithIncludeTags("PROD"))
	if !f.Allow(entry("svc", "prod")) {
		t.Fatal("expected tag matching to be case-insensitive")
	}
}

func TestAllow_CombinedRules_AllMustPass(t *testing.T) {
	f := filter.New(
		filter.WithIncludeTags("prod"),
		filter.WithExcludeTags("legacy"),
		filter.WithPattern("api-*"),
	)
	// all rules satisfied
	if !f.Allow(entry("api-svc", "prod")) {
		t.Fatal("expected entry satisfying all rules to be allowed")
	}
	// include tag missing
	if f.Allow(entry("api-svc", "staging")) {
		t.Fatal("expected entry missing include tag to be rejected")
	}
	// exclude tag present
	if f.Allow(entry("api-svc", "prod", "legacy")) {
		t.Fatal("expected entry with excluded tag to be rejected")
	}
	// pattern mismatch
	if f.Allow(entry("worker", "prod")) {
		t.Fatal("expected entry with non-matching pattern to be rejected")
	}
}
