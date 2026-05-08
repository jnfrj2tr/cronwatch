package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestRateLimitHandler(t *testing.T) *rateLimitHandler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ratelimit.json")
	store := monitor.NewAlertRateLimit(path, 5*time.Minute)
	return newRateLimitHandler(store, []string{"backup", "sync"})
}

func TestHandleRateLimit_GetAllowed(t *testing.T) {
	h := newTestRateLimitHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/ratelimit?job=backup", nil)
	rec := httptest.NewRecorder()
	h.handleRateLimit(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)
	if body["allowed"] != true {
		t.Errorf("expected allowed=true for fresh job")
	}
}

func TestHandleRateLimit_GetAfterRecord(t *testing.T) {
	h := newTestRateLimitHandler(t)
	_ = h.store.Record("backup")
	req := httptest.NewRequest(http.MethodGet, "/ratelimit?job=backup", nil)
	rec := httptest.NewRecorder()
	h.handleRateLimit(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)
	if body["allowed"] != false {
		t.Errorf("expected allowed=false within cooldown")
	}
}

func TestHandleRateLimit_Reset(t *testing.T) {
	h := newTestRateLimitHandler(t)
	_ = h.store.Record("sync")
	req := httptest.NewRequest(http.MethodDelete, "/ratelimit?job=sync", nil)
	rec := httptest.NewRecorder()
	h.handleRateLimit(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if !h.store.Allow("sync") {
		t.Error("expected job to be allowed after reset")
	}
}

func TestHandleRateLimit_UnknownJob(t *testing.T) {
	h := newTestRateLimitHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/ratelimit?job=ghost", nil)
	rec := httptest.NewRecorder()
	h.handleRateLimit(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleRateLimit_MissingJobParam(t *testing.T) {
	h := newTestRateLimitHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/ratelimit", nil)
	rec := httptest.NewRecorder()
	h.handleRateLimit(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleRateLimit_MethodNotAllowed(t *testing.T) {
	h := newTestRateLimitHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/ratelimit?job=backup", nil)
	rec := httptest.NewRecorder()
	h.handleRateLimit(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
