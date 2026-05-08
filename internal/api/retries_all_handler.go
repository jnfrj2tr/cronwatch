package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/monitor"
)

type retriesAllHandler struct {
	store *monitor.RetryStore
	cfg   *config.Config
}

func newRetriesAllHandler(store *monitor.RetryStore, cfg *config.Config) *retriesAllHandler {
	return &retriesAllHandler{store: store, cfg: cfg}
}

// ServeHTTP handles GET /retries/all — returns retry counts for every known job.
func (h *retriesAllHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type entry struct {
		Job   string `json:"job"`
		Count int    `json:"count"`
	}

	results := make([]entry, 0, len(h.cfg.Jobs))
	for _, job := range h.cfg.Jobs {
		count, _ := h.store.Get(job.Name)
		results = append(results, entry{Job: job.Name, Count: count})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
