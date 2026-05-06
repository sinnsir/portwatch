// Package pipeline implements a lightweight, ordered stage executor for
// port-watch events.
//
// Typical usage:
//
//	pl := pipeline.New(
//		alertStage,
//		throttleStage,
//		notifyStage,
//	)
//	if err := pl.Run(ctx, event); err != nil {
//		log.Println(err)
//	}
//
// A Stage may return pipeline.ErrSkip to silently drop the event without
// propagating an error to the caller.
package pipeline
