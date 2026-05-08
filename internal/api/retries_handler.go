package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/user/cronwatch/internal/monitor"
)

type retriesHandler struct {
	store *monitor.RetryStore
}

func newRetriesHandler(store *monitor.RetryStore) *retriesHandler {
	return &retriesHandler{store: store}
}

// ServeHTTP routes GET and DELETE requests for /api/retries?job=<name>.
func (h *retriesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodDelete:
		h.handleReset(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *retriesHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	count, err := h.store.Get(job)
	if err == monitor.ErrUnknownJob {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job":   job,
		"count": count,
	})
}

func (h *retriesHandler) handleReset(w http.ResponseWriter, r *http.Request) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	err := h.store.Reset(job)
	if err == monitor.ErrUnknownJob {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reset", "job": job})
}
