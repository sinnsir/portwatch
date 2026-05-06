package pipeline

import (
	"context"
	"fmt"
)

// FuncStage wraps a plain function so it satisfies the Stage interface.
type FuncStage struct {
	name string
	fn   func(ctx context.Context, e *Event) error
}

// Func creates a Stage from a function.
func Func(name string, fn func(ctx context.Context, e *Event) error) Stage {
	return &FuncStage{name: name, fn: fn}
}

func (f *FuncStage) Name() string { return f.name }
func (f *FuncStage) Process(ctx context.Context, e *Event) error {
	return f.fn(ctx, e)
}

// StateFilter returns a Stage that skips events whose State is not in the
// allowed set.
func StateFilter(allowed ...string) Stage {
	set := make(map[string]struct{}, len(allowed))
	for _, s := range allowed {
		set[s] = struct{}{}
	}
	return Func("state-filter", func(_ context.Context, e *Event) error {
		if _, ok := set[e.State]; !ok {
			return ErrSkip
		}
		return nil
	})
}

// PortFilter returns a Stage that skips events whose Port is not in the
// allowed set.
func PortFilter(ports ...int) Stage {
	set := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		set[p] = struct{}{}
	}
	return Func("port-filter", func(_ context.Context, e *Event) error {
		if _, ok := set[e.Port]; !ok {
			return ErrSkip
		}
		return nil
	})
}

// MetaEnricher returns a Stage that merges the provided key/value pairs
// into Event.Meta, allocating the map if necessary.
func MetaEnricher(kv map[string]string) Stage {
	return Func("meta-enricher", func(_ context.Context, e *Event) error {
		if e.Meta == nil {
			e.Meta = make(map[string]string, len(kv))
		}
		for k, v := range kv {
			e.Meta[k] = v
		}
		return nil
	})
}

// Validator returns a Stage that ensures the event has non-empty Host and
// a positive Port number.
func Validator() Stage {
	return Func("validator", func(_ context.Context, e *Event) error {
		if e.Host == "" {
			return fmt.Errorf("event host is empty")
		}
		if e.Port <= 0 || e.Port > 65535 {
			return fmt.Errorf("event port %d out of range", e.Port)
		}
		return nil
	})
}
