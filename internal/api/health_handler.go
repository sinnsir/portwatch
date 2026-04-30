package api

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/portwatch/internal/healthcheck"
)

// healthHandler wraps a healthcheck.Checker and serves its status over HTTP.
type healthHandler struct {
	checker *healthcheck.Checker
}

// newHealthHandler creates a handler backed by the given Checker.
func newHealthHandler(c *healthcheck.Checker) *healthHandler {
	return &healthHandler{checker: c}
}

// ServeHTTP writes the health status as JSON.
// Responds with 200 OK when healthy, 503 Service Unavailable otherwise.
func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := h.checker.Check()

	code := http.StatusOK
	if !status.Healthy {
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// RegisterHealthHandler mounts the health endpoint on mux at the given path.
func RegisterHealthHandler(mux *http.ServeMux, path string, c *healthcheck.Checker) {
	mux.Handle(path, newHealthHandler(c))
}
