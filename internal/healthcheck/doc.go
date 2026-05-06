// Package healthcheck provides a lightweight self-diagnostic subsystem for
// the portwatch daemon.
//
// A Checker holds a set of named Probe functions. Calling Check() runs every
// probe and returns a Status that summarises the daemon's health, uptime,
// runtime metadata, and per-component results.
//
// # Overview
//
// The package is intentionally minimal: probes are plain functions that return
// an error (nil means healthy), and the resulting Status is a plain struct that
// can be serialised to JSON for an HTTP health endpoint or logged directly.
//
// # Example usage
//
//	probe := healthcheck.Probe{
//		Name: "history_store",
//		Fn: func() error {
//			// verify the store is reachable
//			return nil
//		},
//	}
//	checker := healthcheck.New(probe)
//	status := checker.Check()
//	if !status.Healthy {
//		log.Println("daemon unhealthy:", status.Components)
//	}
//
// The overall Status.Healthy field is true only when every registered probe
// returns a nil error. A single failing probe marks the whole daemon unhealthy.
package healthcheck
