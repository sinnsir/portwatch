// Package sampler provides probabilistic event sampling for the portwatch
// monitoring pipeline.
//
// A Sampler is initialised with a Policy that specifies the fraction of
// events (Rate) that should be allowed through. Calling Allow() on each
// event returns true with probability equal to Rate, making it easy to
// shed load without dropping all traffic.
//
// Example:
//
//	s := sampler.New(sampler.Policy{Rate: 0.25})
//	if s.Allow() {
//		// process ~25 % of events
//	}
//
// All methods are safe for concurrent use.
package sampler
