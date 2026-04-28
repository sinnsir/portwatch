package history

import "time"

// Filter holds optional criteria for querying history events.
type Filter struct {
	// Port restricts results to a specific port. Zero means all ports.
	Port int
	// Since restricts results to events at or after this time.
	Since time.Time
	// State restricts results to a specific state string ("open" or "closed").
	// Empty string means all states.
	State string
	// Limit caps the number of results returned. Zero means no limit.
	Limit int
}

// Query returns a filtered snapshot of recorded events.
// Events are returned in chronological order (oldest first).
func (h *History) Query(f Filter) []Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	all := h.snapshot()

	var result []Event
	for _, e := range all {
		if f.Port != 0 && e.Port != f.Port {
			continue
		}
		if !f.Since.IsZero() && e.Timestamp.Before(f.Since) {
			continue
		}
		if f.State != "" && e.State != f.State {
			continue
		}
		result = append(result, e)
		if f.Limit > 0 && len(result) >= f.Limit {
			break
		}
	}
	return result
}
