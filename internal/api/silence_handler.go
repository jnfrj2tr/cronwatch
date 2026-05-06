package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

type silenceRequest struct {
	JobName  string `json:"job_name"`
	Duration string `json:"duration"`
}

type silenceResponse struct {
	JobName  string    `json:"job_name"`
	Until    time.Time `json:"until"`
	Message  string    `json:"message"`
}

func (h *Handler) handleSilence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req silenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.JobName == "" {
		http.Error(w, "job_name is required", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil || duration <= 0 {
		http.Error(w, "invalid duration: use Go duration format e.g. 2h, 30m", http.StatusBadRequest)
		return
	}

	until, err := h.monitor.Silence(req.JobName, duration)
	if err != nil {
		if err == monitor.ErrUnknownJob {
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(silenceResponse{
		JobName: req.JobName,
		Until:   until,
		Message: "alerts silenced",
	})
}

func (h *Handler) handleUnsilence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobName := r.URL.Query().Get("job")
	if jobName == "" {
		http.Error(w, "job query parameter is required", http.StatusBadRequest)
		return
	}

	if err := h.monitor.Unsilence(jobName); err != nil {
		if err == monitor.ErrUnknownJob {
			http.Error(w, "unknown job", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
