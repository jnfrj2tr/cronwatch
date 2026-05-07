package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cespare/cronwatch/internal/monitor"
)

type runLogHandler struct {
	runLog  *monitor.RunLog
	knownJobs map[string]struct{}
}

func newRunLogHandler(rl *monitor.RunLog, jobs []string) *runLogHandler {
	km := make(map[string]struct{}, len(jobs))
	for _, j := range jobs {
		km[j] = struct{}{}
	}
	return &runLogHandler{runLog: rl, knownJobs: km}
}

// ServeHTTP handles GET /api/runlog?job=<name>&limit=<n>
func (h *runLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobName := r.URL.Query().Get("job")
	if jobName == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}

	if _, ok := h.knownJobs[jobName]; !ok {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}

	limit := 10
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		n, err := strconv.Atoi(lStr)
		if err != nil || n <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = n
	}

	results := h.runLog.Recent(jobName, limit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"job":     jobName,
		"results": results,
	})
}
