package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/user/cronwatch/internal/monitor"
)

const defaultHistoryLimit = 10

type historyProvider interface {
	Recent(jobName string, n int) []monitor.HistoryEntry
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobName := r.URL.Query().Get("job")
	if jobName == "" {
		http.Error(w, "missing required query param: job", http.StatusBadRequest)
		return
	}

	if !h.monitor.JobExists(jobName) {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}

	limit := defaultHistoryLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n <= 0 {
			http.Error(w, "invalid limit: must be a positive integer", http.StatusBadRequest)
			return
		}
		limit = n
	}

	entries := h.history.Recent(jobName, limit)
	if entries == nil {
		entries = []monitor.HistoryEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job":     jobName,
		"entries": entries,
	})
}
