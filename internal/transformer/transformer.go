// Package transformer provides a pipeline stage that mutates or reshapes
// port-state events before they are dispatched to downstream handlers.
package transformer

import (
	"fmt"
	"strings"
	"time"
)

// Event is the minimal contract this package operates on so that it remains
// decoupled from any concrete event type in the wider codebase.
type Event struct {
	Host      string
	Port      int
	State     string
	Timestamp time.Time
	Meta      map[string]string
}

// TransformFunc mutates an Event in-place and returns an error if the
// transformation cannot be applied.
type TransformFunc func(e *Event) error

// Transformer applies an ordered chain of TransformFuncs to every event.
type Transformer struct {
	steps []TransformFunc
}

// New returns a Transformer that will apply steps in the order provided.
func New(steps ...TransformFunc) *Transformer {
	return &Transformer{steps: steps}
}

// Apply runs all registered steps against e, stopping and returning the first
// error encountered.
func (t *Transformer) Apply(e *Event) error {
	for _, fn := range t.steps {
		if err := fn(e); err != nil {
			return err
		}
	}
	return nil
}

// Len returns the number of transformation steps registered.
func (t *Transformer) Len() int { return len(t.steps) }

// --- built-in transforms ---------------------------------------------------

// NormaliseHost lowercases the Host field and trims surrounding whitespace.
func NormaliseHost(e *Event) error {
	e.Host = strings.ToLower(strings.TrimSpace(e.Host))
	return nil
}

// EnsureTimestamp sets Timestamp to time.Now() when it is the zero value.
func EnsureTimestamp(e *Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	return nil
}

// AddMeta returns a TransformFunc that merges the supplied key/value pairs
// into the event's Meta map, creating the map when it is nil.
func AddMeta(pairs map[string]string) TransformFunc {
	return func(e *Event) error {
		if e.Meta == nil {
			e.Meta = make(map[string]string, len(pairs))
		}
		for k, v := range pairs {
			e.Meta[k] = v
		}
		return nil
	}
}

// RequirePort returns an error when the event's Port is outside [1, 65535].
func RequirePort(e *Event) error {
	if e.Port < 1 || e.Port > 65535 {
		return fmt.Errorf("transformer: invalid port %d", e.Port)
	}
	return nil
}
