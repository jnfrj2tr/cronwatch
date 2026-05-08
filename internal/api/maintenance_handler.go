package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

type maintenanceHandler struct {
	store *monitor.MaintenanceStore
}

func newMaintenanceHandler(store *monitor.MaintenanceStore) *maintenanceHandler {
	return &maintenanceHandler{store: store}
}

func (h *maintenanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
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

func (h *maintenanceHandler) handleGet(w http.ResponseWriter, job string) {
	ws, err := h.store.Get(job)
	if err == monitor.ErrUnknownJob {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ws)
}

func (h *maintenanceHandler) handlePost(w http.ResponseWriter, r *http.Request, job string) {
	var body struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		Reason    string    `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.EndTime.IsZero() || !body.EndTime.After(body.StartTime) {
		http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
		return
	}
	win := monitor.MaintenanceWindow{
		JobName:   job,
		StartTime: body.StartTime,
		EndTime:   body.EndTime,
		Reason:    body.Reason,
	}
	if err := h.store.Set(win); err == monitor.ErrUnknownJob {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *maintenanceHandler) handleDelete(w http.ResponseWriter, job string) {
	if err := h.store.Delete(job); err == monitor.ErrUnknownJob {
		http.Error(w, "unknown job", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
