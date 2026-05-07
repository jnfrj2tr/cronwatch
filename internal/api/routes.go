package api

import (
	"net/http"
)

// RegisterRoutes wires all API handlers onto the given mux.
func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/health", h.HandleHealth)
	mux.HandleFunc("/api/heartbeat", h.HandleHeartbeat)
	mux.HandleFunc("/api/status", h.HandleStatus)
	mux.HandleFunc("/api/silence", h.HandleSilence)
	mux.HandleFunc("/api/unsilence", h.HandleUnsilence)
	mux.HandleFunc("/api/history", h.HandleHistory)

	if h.metrics != nil {
		mh := newMetricsHandler(h.metrics, h.knownJobs())
		mux.Handle("/api/metrics", mh)
		meh := newMetricsExportHandler(h.metrics, h.knownJobs())
		mux.Handle("/metrics", meh)
	}

	if h.tags != nil {
		tags := newTagsHandler(h.tags, h.knownJobs())
		mux.Handle("/api/jobs/", tags)
	}
}
