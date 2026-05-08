// Package probe provides a composable HTTP and TCP liveness probe that can be
// wired into the healthcheck subsystem or used standalone.
package probe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Kind identifies the transport used by a Probe.
type Kind string

const (
	KindTCP  Kind = "tcp"
	KindHTTP Kind = "http"
)

// Result holds the outcome of a single probe execution.
type Result struct {
	Kind    Kind
	Target  string
	Healthy bool
	Latency time.Duration
	Err     error
}

// Probe executes a single liveness check against a target.
type Probe struct {
	kind    Kind
	target  string
	timeout time.Duration
	client  *http.Client
}

// New creates a Probe for the given kind and target.
// timeout defaults to 5 s when zero.
func New(kind Kind, target string, timeout time.Duration) *Probe {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Probe{
		kind:    kind,
		target:  target,
		timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Run executes the probe and returns a Result.
func (p *Probe) Run(ctx context.Context) Result {
	start := time.Now()
	var err error

	switch p.kind {
	case KindTCP:
		err = p.probeTCP(ctx)
	case KindHTTP:
		err = p.probeHTTP(ctx)
	default:
		err = fmt.Errorf("unknown probe kind: %s", p.kind)
	}

	return Result{
		Kind:    p.kind,
		Target:  p.target,
		Healthy: err == nil,
		Latency: time.Since(start),
		Err:     err,
	}
}

func (p *Probe) probeTCP(ctx context.Context) error {
	dialer := &net.Dialer{Timeout: p.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", p.target)
	if err != nil {
		return err
	}
	return conn.Close()
}

func (p *Probe) probeHTTP(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.target, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}
