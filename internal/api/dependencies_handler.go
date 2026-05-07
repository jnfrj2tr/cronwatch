package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

type dependenciesHandler struct {
	store *monitor.DependencyStore
}

func newDependenciesHandler(store *monitor.DependencyStore) *dependenciesHandler {
	return &dependenciesHandler{store: store}
}

// GET  /api/dependencies?job=<name>  — get deps for a job
// POST /api/dependencies?job=<name>  — set deps for a job (body: {"deps":[...]})
// DELETE /api/dependencies?job=<name> — clear deps for a job
func (h *dependenciesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		deps, ok := h.store.Get(job)
		if !ok {
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string][]string{"deps": deps})

	case http.MethodPost:
		var body struct {
			Deps []string `json:"deps"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if err := h.store.Set(job, body.Deps); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if err := h.store.Delete(job); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *dependenciesHandler) handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(h.store.All())
}
