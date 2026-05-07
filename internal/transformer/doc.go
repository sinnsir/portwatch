// Package transformer implements composable, ordered transformations for
// port-state events produced by the portwatch monitor.
//
// A [Transformer] holds a chain of [TransformFunc] values that are applied
// sequentially to each event.  Processing stops at the first error, making
// it easy to use validation steps alongside mutation steps in the same chain.
//
// Several built-in transforms are provided:
//
//   - [NormaliseHost]  – lower-cases and trims the host field.
//   - [EnsureTimestamp] – back-fills a zero timestamp with time.Now().
//   - [AddMeta]        – merges arbitrary key/value pairs into Meta.
//   - [RequirePort]    – rejects events with out-of-range port numbers.
//
// Custom transforms can be added by implementing the [TransformFunc] signature.
package transformer
