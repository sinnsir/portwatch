package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/portwatch/internal/metrics"
)

type metricsHandler struct {
	reg *metrics.Registry
}

func newMetricsHandler(reg *metrics.Registry) *metricsHandler {
	return &metricsHandler{reg: reg}
}

// ServeHTTP handles GET /metrics and returns a JSON snapshot of all
// registered counters and gauges.
func (h *metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	snap := h.reg.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(snap); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// RegisterMetricsHandler mounts the metrics endpoint on mux.
func RegisterMetricsHandler(mux *http.ServeMux, reg *metrics.Registry) {
	h := newMetricsHandler(reg)
	mux.Handle("/metrics", h)
}
