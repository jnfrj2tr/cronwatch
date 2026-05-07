package api

import "net/http"

// RegisterRoutes wires all HTTP handlers onto mux.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/heartbeat", h.handleHeartbeat)
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/silence", h.silenceHandler.handleSilence)
	mux.HandleFunc("/unsilence", h.silenceHandler.handleUnsilence)
	mux.HandleFunc("/history", h.handleHistory)
	mux.HandleFunc("/metrics", h.metricsHandler.handleMetrics)
}
