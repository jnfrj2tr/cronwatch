package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestDepsHandler(t *testing.T) *dependenciesHandler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "deps.json")
	store, err := monitor.NewDependencyStore(path, []string{"jobA", "jobB", "jobC"})
	if err != nil {
		t.Fatal(err)
	}
	return newDependenciesHandler(store)
}

func TestHandleDependencies_SetAndGet(t *testing.T) {
	h := newTestDepsHandler(t)

	body, _ := json.Marshal(map[string][]string{"deps": {"jobA", "jobB"}})
	req := httptest.NewRequest(http.MethodPost, "/api/dependencies?job=jobC", bytes.NewReader(body))
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/dependencies?job=jobC", nil)
	rw2 := httptest.NewRecorder()
	h.ServeHTTP(rw2, req2)
	if rw2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw2.Code)
	}
	var resp map[string][]string
	_ = json.NewDecoder(rw2.Body).Decode(&resp)
	if len(resp["deps"]) != 2 {
		t.Fatalf("expected 2 deps, got %v", resp["deps"])
	}
}

func TestHandleDependencies_UnknownJob(t *testing.T) {
	h := newTestDepsHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/dependencies?job=ghost", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rw.Code)
	}
}

func TestHandleDependencies_MissingJobParam(t *testing.T) {
	h := newTestDepsHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/dependencies", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleDependencies_Delete(t *testing.T) {
	h := newTestDepsHandler(t)
	body, _ := json.Marshal(map[string][]string{"deps": {"jobA"}})
	req := httptest.NewRequest(http.MethodPost, "/api/dependencies?job=jobB", bytes.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodDelete, "/api/dependencies?job=jobB", nil)
	rw2 := httptest.NewRecorder()
	h.ServeHTTP(rw2, req2)
	if rw2.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rw2.Code)
	}
}

func TestHandleDependencies_MethodNotAllowed(t *testing.T) {
	h := newTestDepsHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/api/dependencies?job=jobA", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestHandleDependencies_All(t *testing.T) {
	h := newTestDepsHandler(t)
	body, _ := json.Marshal(map[string][]string{"deps": {"jobA"}})
	req := httptest.NewRequest(http.MethodPost, "/api/dependencies?job=jobB", bytes.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodGet, "/api/dependencies/all", nil)
	rw2 := httptest.NewRecorder()
	h.handleAll(rw2, req2)
	if rw2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw2.Code)
	}
	var all map[string][]string
	_ = json.NewDecoder(rw2.Body).Decode(&all)
	if len(all) == 0 {
		t.Fatal("expected at least one dependency entry")
	}
}
