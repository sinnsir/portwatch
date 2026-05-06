package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"portwatch/internal/pipeline"
)

func baseEvent() pipeline.Event {
	return pipeline.Event{Host: "localhost", Port: 8080, State: "open"}
}

func TestPipeline_RunsAllStages(t *testing.T) {
	var visited []string
	record := func(name string) pipeline.Stage {
		return pipeline.Func(name, func(_ context.Context, _ *pipeline.Event) error {
			visited = append(visited, name)
			return nil
		})
	}
	pl := pipeline.New(record("a"), record("b"), record("c"))
	if err := pl.Run(context.Background(), baseEvent()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(visited) != 3 {
		t.Fatalf("expected 3 stages visited, got %d", len(visited))
	}
}

func TestPipeline_StopsOnError(t *testing.T) {
	var second bool
	pl := pipeline.New(
		pipeline.Func("fail", func(_ context.Context, _ *pipeline.Event) error {
			return errors.New("boom")
		}),
		pipeline.Func("second", func(_ context.Context, _ *pipeline.Event) error {
			second = true
			return nil
		}),
	)
	err := pl.Run(context.Background(), baseEvent())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if second {
		t.Fatal("second stage should not have run")
	}
}

func TestPipeline_ErrSkipSilent(t *testing.T) {
	var second bool
	pl := pipeline.New(
		pipeline.Func("skip", func(_ context.Context, _ *pipeline.Event) error {
			return pipeline.ErrSkip
		}),
		pipeline.Func("second", func(_ context.Context, _ *pipeline.Event) error {
			second = true
			return nil
		}),
	)
	if err := pl.Run(context.Background(), baseEvent()); err != nil {
		t.Fatalf("ErrSkip should not surface: %v", err)
	}
	if second {
		t.Fatal("second stage should not have run after ErrSkip")
	}
}

func TestPipeline_Len(t *testing.T) {
	pl := pipeline.New(
		pipeline.Validator(),
		pipeline.StateFilter("open"),
	)
	if pl.Len() != 2 {
		t.Fatalf("expected Len 2, got %d", pl.Len())
	}
}

func TestPipeline_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pl := pipeline.New(
		pipeline.Func("noop", func(_ context.Context, _ *pipeline.Event) error { return nil }),
	)
	if err := pl.Run(ctx, baseEvent()); err == nil {
		t.Fatal("expected context error")
	}
}
