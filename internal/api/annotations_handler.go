package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/cronwatch/internal/monitor"
)

type annotationsHandler struct {
	store   *monitor.AnnotationStore
	knownJobs map[string]struct{}
}

func newAnnotationsHandler(store *monitor.AnnotationStore, jobs []string) *annotationsHandler {
	m := make(map[string]struct{}, len(jobs))
	for _, j := range jobs {
		m[j] = struct{}{}
	}
	return &annotationsHandler{store: store, knownJobs: m}
}

// ServeHTTP routes GET/POST/DELETE on /api/annotations?job=<name>
func (h *annotationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	if _, ok := h.knownJobs[job]; !ok {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, job)
	case http.MethodPost:
		h.handlePost(w, r, job)
	case http.MethodDelete:
		h.handleDelete(w, job)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *annotationsHandler) handleGet(w http.ResponseWriter, job string) {
	a, ok := h.store.Get(job)
	if !ok {
		http.Error(w, "no annotation found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}

func (h *annotationsHandler) handlePost(w http.ResponseWriter, r *http.Request, job string) {
	var body struct {
		Note   string `json:"note"`
		Author string `json:"author"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := h.store.Set(job, body.Note, body.Author); err != nil {
		http.Error(w, "failed to save annotation", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *annotationsHandler) handleDelete(w http.ResponseWriter, job string) {
	if err := h.store.Delete(job); err != nil {
		http.Error(w, "failed to delete annotation", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
