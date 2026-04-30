// Package circuit provides a thread-safe circuit breaker used by the runner
// to protect outbound webhook and shell-command actions from cascading
// failures.
//
// A Breaker starts in the Closed state (calls allowed). After MaxFailures
// consecutive failures it moves to Open (calls blocked). Once the Cooldown
// duration has elapsed it transitions to HalfOpen, allowing a single probe
// call through. A successful probe closes the circuit; another failure
// re-opens it.
package circuit
