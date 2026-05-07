package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatch/internal/config"
	"github.com/user/cronwatch/internal/monitor"
)

func newTestAnnotationsHandler(t *testing.T) http.Handler {
	t.Helper()
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "backup", Schedule: "0 2 * * *"},
		},
	}
	store := monitor.NewAnnotationStore(t.TempDir() + "/annotations.json")
	h := newAnnotationsHandler(cfg, store)
	return h
}

func TestHandleAnnotations_SetAndGet(t *testing.T) {
	h := newTestAnnotationsHandler(t)

	body, _ := json.Marshal(map[string]string{"text": "deployed v1.2"})
	req := httptest.NewRequest(http.MethodPost, "/annotations?job=backup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/annotations?job=backup", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(rec2.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(result))
	}
}

func TestHandleAnnotations_UnknownJob(t *testing.T) {
	h := newTestAnnotationsHandler(t)
	body, _ := json.Marshal(map[string]string{"text": "note"})
	req := httptest.NewRequest(http.MethodPost, "/annotations?job=ghost", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleAnnotations_MissingJobParam(t *testing.T) {
	h := newTestAnnotationsHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/annotations", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleAnnotations_Delete(t *testing.T) {
	h := newTestAnnotationsHandler(t)

	body, _ := json.Marshal(map[string]string{"text": "to delete"})
	req := httptest.NewRequest(http.MethodPost, "/annotations?job=backup", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("set: expected 200, got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/annotations?job=backup", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/annotations?job=backup", nil)
	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, req3)
	var result []map[string]interface{}
	json.NewDecoder(rec3.Body).Decode(&result)
	if len(result) != 0 {
		t.Fatalf("expected 0 annotations after delete, got %d", len(result))
	}
}

func TestHandleAnnotations_MethodNotAllowed(t *testing.T) {
	h := newTestAnnotationsHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/annotations?job=backup", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
