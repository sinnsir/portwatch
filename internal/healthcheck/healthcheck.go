// Package healthcheck provides self-diagnostic reporting for the portwatch daemon.
package healthcheck

import (
	"runtime"
	"time"
)

// Status represents the overall health of the daemon.
type Status struct {
	Healthy    bool              `json:"healthy"`
	Uptime     string            `json:"uptime"`
	StartedAt  time.Time         `json:"started_at"`
	GoVersion  string            `json:"go_version"`
	NumCPU     int               `json:"num_cpu"`
	Goroutines int               `json:"goroutines"`
	Components map[string]string `json:"components"`
}

// Probe is a named health check function that returns an error if unhealthy.
type Probe struct {
	Name string
	Fn   func() error
}

// Checker collects component probes and produces a Status.
type Checker struct {
	startedAt time.Time
	probes    []Probe
}

// New creates a new Checker with the given probes.
func New(probes ...Probe) *Checker {
	return &Checker{
		startedAt: time.Now(),
		probes:    probes,
	}
}

// Check runs all registered probes and returns the aggregated Status.
func (c *Checker) Check() Status {
	components := make(map[string]string, len(c.probes))
	healthy := true

	for _, p := range c.probes {
		if err := p.Fn(); err != nil {
			components[p.Name] = "unhealthy: " + err.Error()
			healthy = false
		} else {
			components[p.Name] = "ok"
		}
	}

	uptime := time.Since(c.startedAt).Round(time.Second).String()

	return Status{
		Healthy:    healthy,
		Uptime:     uptime,
		StartedAt:  c.startedAt,
		GoVersion:  runtime.Version(),
		NumCPU:     runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
		Components: components,
	}
}

// StartedAt returns the time the checker was created.
func (c *Checker) StartedAt() time.Time {
	return c.startedAt
}
