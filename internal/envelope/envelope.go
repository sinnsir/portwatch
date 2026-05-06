// Package envelope wraps outbound event payloads with metadata such as
// schema version, source identifier, and send timestamp so downstream
// consumers can perform versioned deserialization without coupling to
// internal types.
package envelope

import (
	"encoding/json"
	"time"
)

const defaultVersion = "v1"

// Envelope wraps an arbitrary payload with routing and versioning metadata.
type Envelope struct {
	// Version identifies the payload schema, e.g. "v1".
	Version string `json:"version"`
	// Source is the logical name of the component that produced the event.
	Source string `json:"source"`
	// SentAt is the UTC timestamp at which the envelope was created.
	SentAt time.Time `json:"sent_at"`
	// Payload holds the raw JSON of the wrapped event.
	Payload json.RawMessage `json:"payload"`
}

// Option is a functional option for configuring an Envelope.
type Option func(*Envelope)

// WithVersion overrides the default schema version string.
func WithVersion(v string) Option {
	return func(e *Envelope) {
		if v != "" {
			e.Version = v
		}
	}
}

// WithSource sets the source identifier on the envelope.
func WithSource(s string) Option {
	return func(e *Envelope) {
		if s != "" {
			e.Source = s
		}
	}
}

// Wrap serialises payload to JSON and returns an Envelope ready for
// transmission. The caller may supply zero or more Option values to
// override defaults. An error is returned if payload cannot be marshalled.
func Wrap(payload any, opts ...Option) (*Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	e := &Envelope{
		Version: defaultVersion,
		Source:  "portwatch",
		SentAt:  time.Now().UTC(),
		Payload: raw,
	}

	for _, o := range opts {
		o(e)
	}

	return e, nil
}

// Unwrap deserialises the envelope's Payload field into dst.
func (e *Envelope) Unwrap(dst any) error {
	return json.Unmarshal(e.Payload, dst)
}

// MarshalJSON implements json.Marshaler so the envelope can be sent as-is.
func (e *Envelope) MarshalJSON() ([]byte, error) {
	type alias Envelope
	return json.Marshal((*alias)(e))
}
