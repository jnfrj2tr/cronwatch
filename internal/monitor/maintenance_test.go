package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempMaintenancePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "maintenance.json")
}

func TestMaintenanceStore_SetAndGet(t *testing.T) {
	s, err := NewMaintenanceStore(tempMaintenancePath(t), []string{"backup"})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	w := MaintenanceWindow{JobName: "backup", StartTime: now, EndTime: now.Add(time.Hour), Reason: "planned"}
	if err := s.Set(w); err != nil {
		t.Fatal(err)
	}
	ws, err := s.Get("backup")
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 1 || ws[0].Reason != "planned" {
		t.Errorf("unexpected windows: %+v", ws)
	}
}

func TestMaintenanceStore_IsUnderMaintenance(t *testing.T) {
	s, _ := NewMaintenanceStore(tempMaintenancePath(t), []string{"backup"})
	now := time.Now()
	_ = s.Set(MaintenanceWindow{JobName: "backup", StartTime: now.Add(-time.Minute), EndTime: now.Add(time.Hour)})
	if !s.IsUnderMaintenance("backup", now) {
		t.Error("expected job to be under maintenance")
	}
	if s.IsUnderMaintenance("backup", now.Add(2*time.Hour)) {
		t.Error("expected job not under maintenance after window")
	}
}

func TestMaintenanceStore_UnknownJob(t *testing.T) {
	s, _ := NewMaintenanceStore(tempMaintenancePath(t), []string{"backup"})
	err := s.Set(MaintenanceWindow{JobName: "ghost"})
	if err != ErrUnknownJob {
		t.Errorf("expected ErrUnknownJob, got %v", err)
	}
	_, err = s.Get("ghost")
	if err != ErrUnknownJob {
		t.Errorf("expected ErrUnknownJob, got %v", err)
	}
}

func TestMaintenanceStore_Delete(t *testing.T) {
	s, _ := NewMaintenanceStore(tempMaintenancePath(t), []string{"backup"})
	now := time.Now()
	_ = s.Set(MaintenanceWindow{JobName: "backup", StartTime: now, EndTime: now.Add(time.Hour)})
	if err := s.Delete("backup"); err != nil {
		t.Fatal(err)
	}
	ws, _ := s.Get("backup")
	if len(ws) != 0 {
		t.Errorf("expected no windows after delete, got %d", len(ws))
	}
}

func TestMaintenanceStore_PersistsAcrossReload(t *testing.T) {
	path := tempMaintenancePath(t)
	s, _ := NewMaintenanceStore(path, []string{"backup"})
	now := time.Now().Truncate(time.Second)
	_ = s.Set(MaintenanceWindow{JobName: "backup", StartTime: now, EndTime: now.Add(time.Hour), Reason: "reload-test"})

	s2, err := NewMaintenanceStore(path, []string{"backup"})
	if err != nil {
		t.Fatal(err)
	}
	ws, _ := s2.Get("backup")
	if len(ws) != 1 || ws[0].Reason != "reload-test" {
		t.Errorf("expected persisted window, got %+v", ws)
	}
}

func TestMaintenanceStore_BadPath(t *testing.T) {
	_, err := NewMaintenanceStore("/no/such/dir/maintenance.json", []string{"backup"})
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestMaintenanceStore_LoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	s, err := NewMaintenanceStore(path, []string{"backup"})
	if err != nil {
		t.Fatal(err)
	}
	if s.IsUnderMaintenance("backup", time.Now()) {
		t.Error("expected no maintenance on fresh store")
	}
	_ = os.Remove(path)
}
