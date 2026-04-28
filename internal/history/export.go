package history

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

// ExportJSON writes the current history snapshot as a JSON array to w.
func (h *History) ExportJSON(w io.Writer) error {
	events := h.Snapshot()
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(events)
}

// ExportTable writes the current history snapshot as a human-readable
// tab-aligned table to w.
func (h *History) ExportTable(w io.Writer) error {
	events := h.Snapshot()
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TIME\tHOST\tPORT\tSTATE")
	for _, e := range events {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Host,
			e.Port,
			e.State,
		)
	}
	return tw.Flush()
}
