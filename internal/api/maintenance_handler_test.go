package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestMaintenanceHandler(t *testing.T) *maintenanceHandler {
	t.Helper()
	path := filepath.Join(t.TempDir(), "maintenance.json")
	store, err := monitor.NewMaintenanceStore(path, []string{"backup", "sync"})
	if err != nil {
		t.Fatal(err)
	}
	return newMaintenanceHandler(store)
}

func TestHandleMaintenance_SetAndGet(t *testing.T) {
	h := newTestMaintenanceHandler(t)
	now := time.Now().UTC().Truncate(time.Second)
	body, _ := json.Marshal(map[string]interface{}{
		"start_time": now,
		"end_time":   now.Add(time.Hour),
		"reason":     "upgrade",
	})
	req := httptest.NewRequest(http.MethodPost, "/maintenance?job=backup", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/maintenance?job=backup", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
	var ws []monitor.MaintenanceWindow
	if err := json.NewDecoder(rr2.Body).Decode(&ws); err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 || ws[0].Reason != "upgrade" {
		t.Errorf("unexpected windows: %+v", ws)
	}
}

func TestHandleMaintenance_UnknownJob(t *testing.T) {
	h := newTestMaintenanceHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/maintenance?job=ghost", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleMaintenance_MissingJobParam(t *testing.T) {
	h := newTestMaintenanceHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/maintenance", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleMaintenance_Delete(t *testing.T) {
	h := newTestMaintenanceHandler(t)
	now := time.Now().UTC()
	body, _ := json.Marshal(map[string]interface{}{"start_time": now, "end_time": now.Add(time.Hour)})
	req := httptest.NewRequest(http.MethodPost, "/maintenance?job=backup", bytes.NewReader(body))
	h.ServeHTTP(httptest.NewRecorder(), req)

	del := httptest.NewRequest(http.MethodDelete, "/maintenance?job=backup", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, del)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestHandleMaintenance_InvalidBody(t *testing.T) {
	h := newTestMaintenanceHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/maintenance?job=backup", bytes.NewBufferString("not-json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
