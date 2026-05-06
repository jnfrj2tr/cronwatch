package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/api"
	"github.com/yourorg/cronwatch/internal/config"
	"github.com/yourorg/cronwatch/internal/monitor"
)

func makeTestMonitor(t *testing.T) *monitor.Monitor {
	t.Helper()
	cfg := &config.Config{
		CheckInterval: 60,
		Jobs: []config.Job{
			{Name: "backup", Schedule: "0 2 * * *", GracePeriod: 10},
		},
	}
	m, err := monitor.New(cfg, nil)
	if err != nil {
		t.Fatalf("monitor.New: %v", err)
	}
	return m
}

func TestHandleHealth(t *testing.T) {
	h := api.NewHandler(makeTestMonitor(t))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHeartbeat_Success(t *testing.T) {
	h := api.NewHandler(makeTestMonitor(t))
	body, _ := json.Marshal(map[string]string{"job": "backup"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/heartbeat", bytes.NewReader(body))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestHandleHeartbeat_UnknownJob(t *testing.T) {
	h := api.NewHandler(makeTestMonitor(t))
	body, _ := json.Marshal(map[string]string{"job": "nonexistent"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/heartbeat", bytes.NewReader(body))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleStatus(t *testing.T) {
	m := makeTestMonitor(t)
	_ = m.RecordHeartbeat("backup", time.Now())
	h := api.NewHandler(m)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := result["backup"]; !ok {
		t.Error("expected 'backup' in status response")
	}
}
