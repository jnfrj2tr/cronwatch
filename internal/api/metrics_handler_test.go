package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/myorg/cronwatch/internal/monitor"
)

func newPopulatedMetrics() *monitor.MetricsStore {
	s := monitor.NewMetricsStore()
	now := time.Now()
	s.RecordHeartbeat("backup", now, 150*time.Millisecond)
	s.RecordMiss("backup", now.Add(time.Hour))
	return s
}

func TestHandleMetrics_AllJobs(t *testing.T) {
	h := newMetricsHandler(newPopulatedMetrics(), []string{"backup"})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []monitor.JobMetrics
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 job, got %d", len(result))
	}
}

func TestHandleMetrics_SingleJob(t *testing.T) {
	h := newMetricsHandler(newPopulatedMetrics(), []string{"backup"})
	req := httptest.NewRequest(http.MethodGet, "/metrics?job=backup", nil)
	w := httptest.NewRecorder()
	h.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var m monitor.JobMetrics
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if m.JobName != "backup" {
		t.Errorf("expected job 'backup', got %q", m.JobName)
	}
	if m.TotalHeartbeats != 1 {
		t.Errorf("expected 1 heartbeat, got %d", m.TotalHeartbeats)
	}
}

func TestHandleMetrics_UnknownJob(t *testing.T) {
	h := newMetricsHandler(newPopulatedMetrics(), []string{"backup"})
	req := httptest.NewRequest(http.MethodGet, "/metrics?job=ghost", nil)
	w := httptest.NewRecorder()
	h.handleMetrics(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleMetrics_KnownJobNoActivity(t *testing.T) {
	h := newMetricsHandler(monitor.NewMetricsStore(), []string{"nightly"})
	req := httptest.NewRequest(http.MethodGet, "/metrics?job=nightly", nil)
	w := httptest.NewRecorder()
	h.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var m monitor.JobMetrics
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if m.TotalHeartbeats != 0 || m.TotalMisses != 0 {
		t.Error("expected zeroed metrics for job with no activity")
	}
}

func TestHandleMetrics_MethodNotAllowed(t *testing.T) {
	h := newMetricsHandler(monitor.NewMetricsStore(), []string{})
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	w := httptest.NewRecorder()
	h.handleMetrics(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
