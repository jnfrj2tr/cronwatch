package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

// Handler serves the HTTP API for cronwatch.
type Handler struct {
	monitor *monitor.Monitor
	mux     *http.ServeMux
}

// NewHandler creates a new Handler wired to the given monitor.
func NewHandler(m *monitor.Monitor) *Handler {
	h := &Handler{monitor: m, mux: http.NewServeMux()}
	h.mux.HandleFunc("/healthz", h.handleHealth)
	h.mux.HandleFunc("/heartbeat", h.handleHeartbeat)
	h.mux.HandleFunc("/status", h.handleStatus)
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleHealth returns 200 OK to indicate the daemon is alive.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleHeartbeat accepts a POST with a job name and records a heartbeat.
func (h *Handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job string `json:"job"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.monitor.RecordHeartbeat(req.Job, time.Now()); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

// handleStatus returns the current state of all monitored jobs.
func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status := h.monitor.Status()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
