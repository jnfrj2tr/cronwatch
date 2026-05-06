package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSilence_Success(t *testing.T) {
	m := makeTestMonitor()
	h := NewHandler(m)

	body := `{"job_name":"backup","duration":"2h"}`
	req := httptest.NewRequest(http.MethodPost, "/silence", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.handleSilence(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	var resp silenceResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.JobName != "backup" {
		t.Errorf("expected job_name=backup, got %s", resp.JobName)
	}
	if resp.Until.IsZero() {
		t.Error("expected non-zero until time")
	}
}

func TestHandleSilence_UnknownJob(t *testing.T) {
	m := makeTestMonitor()
	h := NewHandler(m)

	body := `{"job_name":"nonexistent","duration":"1h"}`
	req := httptest.NewRequest(http.MethodPost, "/silence", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.handleSilence(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleSilence_InvalidDuration(t *testing.T) {
	m := makeTestMonitor()
	h := NewHandler(m)

	body := `{"job_name":"backup","duration":"notaduration"}`
	req := httptest.NewRequest(http.MethodPost, "/silence", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.handleSilence(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleUnsilence_Success(t *testing.T) {
	m := makeTestMonitor()
	h := NewHandler(m)

	req := httptest.NewRequest(http.MethodDelete, "/silence?job=backup", nil)
	rec := httptest.NewRecorder()

	h.handleUnsilence(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHandleUnsilence_MissingJob(t *testing.T) {
	m := makeTestMonitor()
	h := NewHandler(m)

	req := httptest.NewRequest(http.MethodDelete, "/silence", nil)
	rec := httptest.NewRecorder()

	h.handleUnsilence(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
