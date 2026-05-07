package tagger_test

import (
	"testing"

	"github.com/user/portwatch/internal/portcheck"
	"github.com/user/portwatch/internal/tagger"
)

func baseEvent(port int, state portcheck.State, tags ...string) tagger.Event {
	return tagger.Event{Port: port, State: state, Tags: tags}
}

func TestApply_NoRules_ReturnsSameTags(t *testing.T) {
	tr := tagger.New(nil)
	e := baseEvent(8080, portcheck.StateOpen, "existing")
	got := tr.Apply(e)
	if len(got.Tags) != 1 || got.Tags[0] != "existing" {
		t.Fatalf("expected [existing], got %v", got.Tags)
	}
}

func TestApply_MatchingPortAndState_AddsTags(t *testing.T) {
	tr := tagger.New([]tagger.Rule{
		{Ports: []int{443}, States: []portcheck.State{portcheck.StateOpen}, Tags: []string{"tls", "web"}},
	})
	e := baseEvent(443, portcheck.StateOpen)
	got := tr.Apply(e)
	if !containsAll(got.Tags, "tls", "web") {
		t.Fatalf("expected tls and web tags, got %v", got.Tags)
	}
}

func TestApply_NonMatchingPort_SkipsRule(t *testing.T) {
	tr := tagger.New([]tagger.Rule{
		{Ports: []int{443}, Tags: []string{"tls"}},
	})
	got := tr.Apply(baseEvent(80, portcheck.StateOpen))
	if contains(got.Tags, "tls") {
		t.Fatal("expected tls tag to be absent")
	}
}

func TestApply_NonMatchingState_SkipsRule(t *testing.T) {
	tr := tagger.New([]tagger.Rule{
		{States: []portcheck.State{portcheck.StateOpen}, Tags: []string{"up"}},
	})
	got := tr.Apply(baseEvent(80, portcheck.StateClosed))
	if contains(got.Tags, "up") {
		t.Fatal("expected up tag to be absent")
	}
}

func TestApply_EmptyPortsAndStates_MatchesAll(t *testing.T) {
	tr := tagger.New([]tagger.Rule{
		{Tags: []string{"any"}},
	})
	for _, state := range []portcheck.State{portcheck.StateOpen, portcheck.StateClosed} {
		got := tr.Apply(baseEvent(9999, state))
		if !contains(got.Tags, "any") {
			t.Fatalf("expected any tag for state %v, got %v", state, got.Tags)
		}
	}
}

func TestApply_DuplicateTagsDeduped(t *testing.T) {
	tr := tagger.New([]tagger.Rule{
		{Tags: []string{"dup"}},
		{Tags: []string{"dup"}},
	})
	got := tr.Apply(baseEvent(80, portcheck.StateOpen))
	count := 0
	for _, tag := range got.Tags {
		if tag == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected dup tag exactly once, got %d times in %v", count, got.Tags)
	}
}

func TestAddRule_AppliedOnNextCall(t *testing.T) {
	tr := tagger.New(nil)
	tr.AddRule(tagger.Rule{Tags: []string{"dynamic"}})
	got := tr.Apply(baseEvent(22, portcheck.StateOpen))
	if !contains(got.Tags, "dynamic") {
		t.Fatalf("expected dynamic tag, got %v", got.Tags)
	}
}

// helpers

func contains(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func containsAll(tags []string, want ...string) bool {
	for _, w := range want {
		if !contains(tags, w) {
			return false
		}
	}
	return true
}
