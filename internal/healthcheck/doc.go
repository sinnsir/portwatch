// Package healthcheck provides a lightweight self-diagnostic subsystem for
// the portwatch daemon.
//
// A Checker holds a set of named Probe functions. Calling Check() runs every
// probe and returns a Status that summarises the daemon's health, uptime,
// runtime metadata, and per-component results.
//
// Example usage:
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
package healthcheck
