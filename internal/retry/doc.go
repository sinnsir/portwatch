// Package retry implements a generic retry mechanism with exponential backoff.
//
// Usage:
//
//	err := retry.Do(ctx, retry.DefaultPolicy(), func() error {
//		return callExternalService()
//	})
//
// The Policy struct controls the number of attempts, initial delay, backoff
// multiplier, and maximum delay cap. retry.ErrExhausted is wrapped into the
// returned error when all attempts fail, allowing callers to distinguish a
// retry-exhaustion from a context cancellation.
package retry
