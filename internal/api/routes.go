package api

import "net/http"

// RegisterRoutes wires all HTTP handlers onto the given mux.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/heartbeat", h.handleHeartbeat)
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/silence", h.handleSilence)
	mux.HandleFunc("/unsilence", h.handleUnsilence)
	mux.HandleFunc("/history", h.handleHistory)
}
