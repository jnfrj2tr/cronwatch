package api

import "net/http"

// RegisterRoutes attaches all API routes to the given mux.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	// Health check
	mux.HandleFunc("/health", h.handleHealth)

	// Heartbeat endpoint — cron jobs ping this to report successful execution
	mux.HandleFunc("/heartbeat", h.handleHeartbeat)

	// Status overview of all monitored jobs
	mux.HandleFunc("/status", h.handleStatus)

	// Silence management — suppress alerts for a job for a given duration
	mux.HandleFunc("/silence", h.routeSilence)
}

// routeSilence dispatches POST and DELETE to the appropriate silence handlers.
func (h *Handler) routeSilence(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleSilence(w, r)
	case http.MethodDelete:
		h.handleUnsilence(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
