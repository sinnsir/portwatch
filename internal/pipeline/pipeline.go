// Package pipeline provides a sequential stage executor that chains
// port-event processors (notifier, alert filter, throttle, etc.) into
// a single composable unit.
package pipeline

import (
	"context"
	"fmt"
)

// Event carries the data flowing through the pipeline.
type Event struct {
	Host  string
	Port  int
	State string // "open" | "closed"
	Meta  map[string]string
}

// Stage is a single processing step.  It may mutate the event or return
// an error to abort the pipeline.  Returning ErrSkip silently stops
// further processing without recording an error.
type Stage interface {
	Name() string
	Process(ctx context.Context, e *Event) error
}

// ErrSkip signals that the event should be dropped without error.
var ErrSkip = fmt.Errorf("pipeline: event skipped")

// Pipeline runs a list of Stage values in order.
type Pipeline struct {
	stages []Stage
}

// New creates a Pipeline from the provided stages.
func New(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Run executes every stage in sequence.  It stops on the first error or
// ErrSkip and returns the stage name alongside any non-skip error.
func (p *Pipeline) Run(ctx context.Context, e Event) error {
	for _, s := range p.stages {
		if err := s.Process(ctx, &e); err != nil {
			if err == ErrSkip {
				return nil
			}
			return fmt.Errorf("stage %q: %w", s.Name(), err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	return nil
}

// Len returns the number of stages registered in the pipeline.
func (p *Pipeline) Len() int { return len(p.stages) }
