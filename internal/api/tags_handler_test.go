package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/svc/cronwatch/internal/monitor"
)

func newTestTagsHandler(t *testing.T) *tagsHandler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tags.json")
	store := monitor.NewTagStore(path)
	return newTagsHandler(store, []string{"backup", "cleanup"})
}

func TestHandleTags_GetEmpty(t *testing.T) {
	h := newTestTagsHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/backup/tags", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "{") {
		t.Error("expected JSON object in response")
	}
}

func TestHandleTags_SetAndGet(t *testing.T) {
	h := newTestTagsHandler(t)
	body := bytes.NewBufferString(`{"key":"env","value":"staging"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/jobs/backup/tags", body)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/jobs/backup/tags", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if !strings.Contains(rr2.Body.String(), "staging") {
		t.Error("expected tag value in response")
	}
}

func TestHandleTags_Delete(t *testing.T) {
	h := newTestTagsHandler(t)
	h.store.Set("backup", "team", "ops")

	req := httptest.NewRequest(http.MethodDelete, "/api/jobs/backup/tags?key=team", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
	if _, ok := h.store.Get("backup")["team"]; ok {
		t.Error("expected tag to be deleted")
	}
}

func TestHandleTags_UnknownJob(t *testing.T) {
	h := newTestTagsHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/ghost/tags", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleTags_MissingKeyOnDelete(t *testing.T) {
	h := newTestTagsHandler(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/jobs/backup/tags", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleTags_MethodNotAllowed(t *testing.T) {
	h := newTestTagsHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/api/jobs/backup/tags", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
