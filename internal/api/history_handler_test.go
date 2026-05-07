package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

type stubHistory struct {
	entries map[string][]monitor.HistoryEntry
}

func (s *stubHistory) Recent(jobName string, n int) []monitor.HistoryEntry {
	all := s.entries[jobName]
	if len(all) <= n {
		return all
	}
	return all[len(all)-n:]
}

func TestHandleHistory_Success(t *testing.T) {
	m := makeTestMonitor()
	now := time.Now()
	h := &Handler{
		monitor: m,
		history: &stubHistory{
			entries: map[string][]monitor.HistoryEntry{
				"backup": {
					{JobName: "backup", ReceivedAt: now.Add(-2 * time.Minute), Status: "ok"},
					{JobName: "backup", ReceivedAt: now.Add(-1 * time.Minute), Status: "ok"},
				},
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/history?job=backup", nil)
	rr := httptest.NewRecorder()
	h.handleHistory(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	entries := resp["entries"].([]interface{})
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHandleHistory_UnknownJob(t *testing.T) {
	m := makeTestMonitor()
	h := &Handler{monitor: m, history: &stubHistory{entries: map[string][]monitor.HistoryEntry{}}}

	req := httptest.NewRequest(http.MethodGet, "/history?job=nonexistent", nil)
	rr := httptest.NewRecorder()
	h.handleHistory(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleHistory_MissingJobParam(t *testing.T) {
	m := makeTestMonitor()
	h := &Handler{monitor: m, history: &stubHistory{entries: map[string][]monitor.HistoryEntry{}}}

	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	rr := httptest.NewRecorder()
	h.handleHistory(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleHistory_InvalidLimit(t *testing.T) {
	m := makeTestMonitor()
	h := &Handler{monitor: m, history: &stubHistory{entries: map[string][]monitor.HistoryEntry{}}}

	req := httptest.NewRequest(http.MethodGet, "/history?job=backup&limit=abc", nil)
	rr := httptest.NewRecorder()
	h.handleHistory(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
