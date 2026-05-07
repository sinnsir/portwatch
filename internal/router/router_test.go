package router_test

import (
	"context"
	"errors"
	"testing"

	"github.com/user/portwatch/internal/router"
)

// stubEvent implements router.Event.
type stubEvent struct {
	port  int
	state string
}

func (s stubEvent) Port() int    { return s.port }
func (s stubEvent) State() string { return s.state }

func TestDispatch_NoRules_ReturnsNil(t *testing.T) {
	r := router.New()
	err := r.Dispatch(context.Background(), stubEvent{port: 8080, state: "open"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDispatch_MatchingRuleInvoked(t *testing.T) {
	r := router.New()
	called := false
	r.Register(router.Rule{
		Match:   func(e router.Event) bool { return e.State() == "open" },
		Handler: func(_ context.Context, _ router.Event) error { called = true; return nil },
	})
	_ = r.Dispatch(context.Background(), stubEvent{port: 80, state: "open"})
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestDispatch_NonMatchingRuleSkipped(t *testing.T) {
	r := router.New()
	called := false
	r.Register(router.Rule{
		Match:   func(e router.Event) bool { return e.State() == "closed" },
		Handler: func(_ context.Context, _ router.Event) error { called = true; return nil },
	})
	_ = r.Dispatch(context.Background(), stubEvent{port: 80, state: "open"})
	if called {
		t.Fatal("expected handler to be skipped")
	}
}

func TestDispatch_StopsOnFirstError(t *testing.T) {
	r := router.New()
	sentinel := errors.New("handler error")
	secondCalled := false

	r.Register(router.Rule{
		Handler: func(_ context.Context, _ router.Event) error { return sentinel },
	})
	r.Register(router.Rule{
		Handler: func(_ context.Context, _ router.Event) error { secondCalled = true; return nil },
	})

	err := r.Dispatch(context.Background(), stubEvent{port: 443, state: "open"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if secondCalled {
		t.Fatal("second handler should not have been called")
	}
}

func TestDispatch_NilMatchAlwaysFires(t *testing.T) {
	r := router.New()
	count := 0
	r.Register(router.Rule{
		Match:   nil,
		Handler: func(_ context.Context, _ router.Event) error { count++; return nil },
	})
	_ = r.Dispatch(context.Background(), stubEvent{port: 22, state: "closed"})
	if count != 1 {
		t.Fatalf("expected 1 invocation, got %d", count)
	}
}

func TestLen_ReflectsRegisteredRules(t *testing.T) {
	r := router.New()
	if r.Len() != 0 {
		t.Fatalf("expected 0, got %d", r.Len())
	}
	r.Register(router.Rule{Handler: func(_ context.Context, _ router.Event) error { return nil }})
	r.Register(router.Rule{Handler: func(_ context.Context, _ router.Event) error { return nil }})
	if r.Len() != 2 {
		t.Fatalf("expected 2, got %d", r.Len())
	}
}
