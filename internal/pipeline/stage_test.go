package pipeline_test

import (
	"context"
	"testing"

	"portwatch/internal/pipeline"
)

func TestStateFilter_AllowsMatchingState(t *testing.T) {
	pl := pipeline.New(pipeline.StateFilter("open"))
	e := pipeline.Event{Host: "h", Port: 80, State: "open"}
	if err := pl.Run(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStateFilter_SkipsNonMatchingState(t *testing.T) {
	var ran bool
	pl := pipeline.New(
		pipeline.StateFilter("open"),
		pipeline.Func("check", func(_ context.Context, _ *pipeline.Event) error {
			ran = true
			return nil
		}),
	)
	e := pipeline.Event{Host: "h", Port: 80, State: "closed"}
	if err := pl.Run(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ran {
		t.Fatal("downstream stage should not have run")
	}
}

func TestPortFilter_SkipsUnwantedPort(t *testing.T) {
	var ran bool
	pl := pipeline.New(
		pipeline.PortFilter(443),
		pipeline.Func("check", func(_ context.Context, _ *pipeline.Event) error {
			ran = true
			return nil
		}),
	)
	e := pipeline.Event{Host: "h", Port: 80, State: "open"}
	if err := pl.Run(context.Background(), e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ran {
		t.Fatal("downstream stage should not have run for filtered port")
	}
}

func TestMetaEnricher_AddsMeta(t *testing.T) {
	var got *pipeline.Event
	pl := pipeline.New(
		pipeline.MetaEnricher(map[string]string{"env": "prod"}),
		pipeline.Func("capture", func(_ context.Context, e *pipeline.Event) error {
			got = e
			return nil
		}),
	)
	if err := pl.Run(context.Background(), baseEvent()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "prod" {
		t.Fatalf("expected meta env=prod, got %v", got.Meta)
	}
}

func TestValidator_RejectsEmptyHost(t *testing.T) {
	pl := pipeline.New(pipeline.Validator())
	e := pipeline.Event{Host: "", Port: 80, State: "open"}
	if err := pl.Run(context.Background(), e); err == nil {
		t.Fatal("expected validation error for empty host")
	}
}

func TestValidator_RejectsInvalidPort(t *testing.T) {
	pl := pipeline.New(pipeline.Validator())
	e := pipeline.Event{Host: "h", Port: 0, State: "open"}
	if err := pl.Run(context.Background(), e); err == nil {
		t.Fatal("expected validation error for port 0")
	}
}
