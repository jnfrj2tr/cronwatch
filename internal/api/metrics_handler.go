package api

import (
	"encoding/json"
	"net/http"

	"github.com/myorg/cronwatch/internal/monitor"
)

// metricsProvider is satisfied by *monitor.MetricsStore.
type metricsProvider interface {
	Get(job string) (monitor.JobMetrics, bool)
	All() []monitor.JobMetrics
}

type metricsHandler struct {
	metrics metricsProvider
	jobs    []string // known job names from config
}

func newMetricsHandler(m metricsProvider, jobs []string) *metricsHandler {
	return &metricsHandler{metrics: m, jobs: jobs}
}

// handleMetrics handles GET /metrics?job=<name>
// Returns all jobs when no query param is supplied.
func (h *metricsHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	job := r.URL.Query().Get("job")
	w.Header().Set("Content-Type", "application/json")

	if job != "" {
		m, ok := h.metrics.Get(job)
		if !ok {
			// Return zeroed metrics for known jobs that have no activity yet.
			known := false
			for _, j := range h.jobs {
				if j == job {
					known = true
					break
				}
			}
			if !known {
				http.Error(w, `{"error":"unknown job"}`, http.StatusNotFound)
				return
			}
			m = monitor.JobMetrics{JobName: job}
		}
		json.NewEncoder(w).Encode(m)
		return
	}

	all := h.metrics.All()
	if all == nil {
		all = []monitor.JobMetrics{}
	}
	json.NewEncoder(w).Encode(all)
}
