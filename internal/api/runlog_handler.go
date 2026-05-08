package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/user/cronwatch/internal/monitor"
)

type runLogHandler struct {
	runLog   *monitor.RunLog
	knownJobs map[string]struct{}
}

func newRunLogHandler(rl *monitor.RunLog, jobs []string) *runLogHandler {
	known := make(map[string]struct{}, len(jobs))
	for _, j := range jobs {
		known[j] = struct{}{}
	}
	return &runLogHandler{runLog: rl, knownJobs: known}
}

func (h *runLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}

	if _, ok := h.knownJobs[job]; !ok {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
	}

	entries := h.runLog.Recent(job, limit)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}
