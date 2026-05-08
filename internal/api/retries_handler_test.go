package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/monitor"
)

func newTestRetriesHandler(t *testing.T) (*retriesHandler, *monitor.RetryStore) {
	t.Helper()
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "backup", Schedule: "@daily"},
			{Name: "sync", Schedule: "@hourly"},
		},
	}
	store, err := monitor.NewRetryStore("", cfg)
	if err != nil {
		t.Fatalf("NewRetryStore: %v", err)
	}
	return newRetriesHandler(store, cfg), store
}

func TestHandleRetries_GetZero(t *testing.T) {
	h, _ := newTestRetriesHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/retries?job=backup", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rw.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["job"] != "backup" {
		t.Errorf("unexpected job: %v", body["job"])
	}
	if int(body["count"].(float64)) != 0 {
		t.Errorf("expected count 0, got %v", body["count"])
	}
}

func TestHandleRetries_UnknownJob(t *testing.T) {
	h, _ := newTestRetriesHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/retries?job=ghost", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rw.Code)
	}
}

func TestHandleRetries_MissingJobParam(t *testing.T) {
	h, _ := newTestRetriesHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/retries", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleRetries_IncrementAndReset(t *testing.T) {
	h, store := newTestRetriesHandler(t)

	store.Increment("sync")
	store.Increment("sync")

	req := httptest.NewRequest(http.MethodGet, "/retries?job=sync", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rw.Body).Decode(&body)
	if int(body["count"].(float64)) != 2 {
		t.Errorf("expected count 2, got %v", body["count"])
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/retries?job=sync", nil)
	rw2 := httptest.NewRecorder()
	h.ServeHTTP(rw2, req2)
	if rw2.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/retries?job=sync", nil)
	rw3 := httptest.NewRecorder()
	h.ServeHTTP(rw3, req3)
	var body2 map[string]interface{}
	json.NewDecoder(rw3.Body).Decode(&body2)
	if int(body2["count"].(float64)) != 0 {
		t.Errorf("expected count 0 after reset, got %v", body2["count"])
	}
}

func TestHandleRetries_MethodNotAllowed(t *testing.T) {
	h, _ := newTestRetriesHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/retries?job=backup", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
