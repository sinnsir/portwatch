package eventlog

import "time"

// Filter holds optional predicates for querying the log.
type Filter struct {
	Port  int       // zero means any port
	State string    // empty means any state
	Since time.Time // zero means no lower bound
	Limit int       // zero means no limit
}

// Query applies f to the entries returned by l.Snapshot and returns the
// matching subset. Results preserve the original oldest-first order.
func (l *Log) Query(f Filter) []Entry {
	all := l.Snapshot()
	out := make([]Entry, 0, len(all))
	for _, e := range all {
		if f.Port != 0 && e.Port != f.Port {
			continue
		}
		if f.State != "" && e.State != f.State {
			continue
		}
		if !f.Since.IsZero() && e.Timestamp.Before(f.Since) {
			continue
		}
		out = append(out, e)
		if f.Limit > 0 && len(out) >= f.Limit {
			break
		}
	}
	return out
}
