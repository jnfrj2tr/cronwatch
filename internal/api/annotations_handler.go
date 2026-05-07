package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/monitor"
)

type annotationsHandler struct {
	cfg   *config.Config
	store *monitor.AnnotationStore
}

func newAnnotationsHandler(cfg *config.Config, store *monitor.AnnotationStore) http.Handler {
	return &annotationsHandler{cfg: cfg, store: store}
}

func (h *annotationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	if !h.jobExists(job) {
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
	annotations := h.store.Get(job)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(annotations)
}

func (h *annotationsHandler) handlePost(w http.ResponseWriter, r *http.Request, job string) {
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	h.store.Add(job, body.Text, time.Now())
	w.WriteHeader(http.StatusOK)
}

func (h *annotationsHandler) handleDelete(w http.ResponseWriter, job string) {
	h.store.Clear(job)
	w.WriteHeader(http.StatusOK)
}

func (h *annotationsHandler) jobExists(name string) bool {
	for _, j := range h.cfg.Jobs {
		if j.Name == name {
			return true
		}
	}
	return false
}
