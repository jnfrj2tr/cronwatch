package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

type rateLimitHandler struct {
	store    *monitor.AlertRateLimit
	knownJobs map[string]bool
}

func newRateLimitHandler(store *monitor.AlertRateLimit, knownJobs []string) *rateLimitHandler {
	jobSet := make(map[string]bool, len(knownJobs))
	for _, j := range knownJobs {
		jobSet[j] = true
	}
	return &rateLimitHandler{store: store, knownJobs: jobSet}
}

// handleRateLimit handles GET (check) and DELETE (reset) for a job's alert rate limit.
func (h *rateLimitHandler) handleRateLimit(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	if !h.knownJobs[job] {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		allowed := h.store.Allow(job)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"job":     job,
			"allowed": allowed,
			"checked": time.Now().UTC(),
		})

	case http.MethodDelete:
		if err := h.store.Reset(job); err != nil {
			http.Error(w, "failed to reset rate limit", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
