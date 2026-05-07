package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/svc/cronwatch/internal/monitor"
)

type tagsHandler struct {
	store  *monitor.TagStore
	jobs   map[string]struct{}
}

func newTagsHandler(store *monitor.TagStore, knownJobs []string) *tagsHandler {
	m := make(map[string]struct{}, len(knownJobs))
	for _, j := range knownJobs {
		m[j] = struct{}{}
	}
	return &tagsHandler{store: store, jobs: m}
}

// GET /api/jobs/{job}/tags
// POST /api/jobs/{job}/tags  body: {"key":"k","value":"v"}
// DELETE /api/jobs/{job}/tags?key=k
func (h *tagsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	job = strings.TrimSuffix(job, "/tags")

	if _, ok := h.jobs[job]; !ok {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(h.store.Get(job))

	case http.MethodPost:
		var body struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		h.store.Set(job, body.Key, body.Value)
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key param", http.StatusBadRequest)
			return
		}
		h.store.Delete(job, key)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
