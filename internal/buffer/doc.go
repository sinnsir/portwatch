// Package buffer implements a generic, bounded event buffer for portwatch.
//
// Items are accumulated in memory and delivered to a FlushFunc in batches.
// A flush is triggered automatically when:
//
//   - the buffer reaches its configured Capacity, or
//   - the FlushInterval elapses (whichever comes first).
//
// The zero Policy is unsafe; use DefaultPolicy or supply explicit values.
// Buffer is safe for concurrent use.
package buffer
