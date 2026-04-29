// Package api provides a lightweight HTTP server exposing portwatch
// runtime state: history snapshots and current port statuses.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/user/portwatch/internal/history"
)

// HistoryStore is the subset of history.History used by the server.
type HistoryStore interface {
	Snapshot() []history.Event
}

// Server is a minimal HTTP API server.
type Server struct {
	addr    string
	history HistoryStore
	mux     *http.ServeMux
	httpSrv *http.Server
}

// New creates a Server bound to addr.
func New(addr string, h HistoryStore) *Server {
	s := &Server{
		addr:    addr,
		history: h,
		mux:     http.NewServeMux(),
	}
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/api/v1/events", s.handleEvents)
	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

// Start begins listening in a goroutine. It returns immediately.
func (s *Server) Start() error {
	ln, err := s.httpSrv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	_ = ln
	return nil
}

// ListenAndServe blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return s.httpSrv.ListenAndServe()
}

// Close shuts the HTTP server down.
func (s *Server) Close() error {
	return s.httpSrv.Close()
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	events := s.history.Snapshot()

	// optional ?limit= query param
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if n, err := strconv.Atoi(lStr); err == nil && n > 0 && n < len(events) {
			events = events[len(events)-n:]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}
