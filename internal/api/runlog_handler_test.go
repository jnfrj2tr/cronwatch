package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestRunLogHandler(t *testing.T, jobs []string) *runLogHandler {
	t.Helper()
	rl := monitor.NewRunLog(t.TempDir()+"/runlog.json", jobs, 50)
	return newRunLogHandler(rl, jobs)
}

func TestHandleRunLog_Success(t *testing.T) {
	jobs := []string{"backup", "sync"}
	h := newTestRunLogHandler(t, jobs)

	now := time.Now()
	_ = h.runLog.Record("backup", true, "completed ok", now)
	_ = h.runLog.Record("backup", false, "exit code 1", now.Add(time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/api/runlog?job=backup", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []monitor.RunEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleRunLog_UnknownJob(t *testing.T) {
	h := newTestRunLogHandler(t, []string{"backup"})

	req := httptest.NewRequest(http.MethodGet, "/api/runlog?job=ghost", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleRunLog_MissingJobParam(t *testing.T) {
	h := newTestRunLogHandler(t, []string{"backup"})

	req := httptest.NewRequest(http.MethodGet, "/api/runlog", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleRunLog_LimitParam(t *testing.T) {
	jobs := []string{"cleanup"}
	h := newTestRunLogHandler(t, jobs)

	now := time.Now()
	for i := 0; i < 10; i++ {
		_ = h.runLog.Record("cleanup", true, "ok", now.Add(time.Duration(i)*time.Minute))
	}

	req := httptest.NewRequest(http.MethodGet, "/api/runlog?job=cleanup&limit=3", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []monitor.RunEntry
	if err := json.NewDecoder(w.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
}

func TestHandleRunLog_MethodNotAllowed(t *testing.T) {
	h := newTestRunLogHandler(t, []string{"backup"})

	req := httptest.NewRequest(http.MethodPost, "/api/runlog?job=backup", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
