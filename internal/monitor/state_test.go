package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadState(t *testing.T) {
	cfg := makeConfig(30)
	m := New(cfg, nil)

	now := time.Now().Truncate(time.Second)
	m.RecordHeartbeat("test-job", now)
	m.status["test-job"].Missed = true
	m.status["test-job"].Consecutive = 3

	tmp := filepath.Join(t.TempDir(), "state.json")
	if err := m.SaveState(tmp); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	// Load into a fresh monitor
	m2 := New(cfg, nil)
	if err := m2.LoadState(tmp); err != nil {
		t.Fatalf("LoadState: %v", err)
	}

	m2.mu.Lock()
	defer m2.mu.Unlock()
	s := m2.status["test-job"]
	if !s.LastSeen.Equal(now) {
		t.Errorf("LastSeen: want %v, got %v", now, s.LastSeen)
	}
	if !s.Missed {
		t.Error("expected Missed=true")
	}
	if s.Consecutive != 3 {
		t.Errorf("Consecutive: want 3, got %d", s.Consecutive)
	}
}

func TestLoadState_MissingFile(t *testing.T) {
	m := New(makeConfig(0), nil)
	err := m.LoadState("/nonexistent/path/state.json")
	if err != nil {
		t.Errorf("expected nil error for missing file, got %v", err)
	}
}

func TestLoadState_UnknownJobIgnored(t *testing.T) {
	cfg := makeConfig(0)
	m := New(cfg, nil)

	// Save state with a known job
	tmp := filepath.Join(t.TempDir(), "state.json")
	if err := m.SaveState(tmp); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	// Load into monitor with different jobs — should not panic
	cfg2 := makeConfig(0)
	cfg2.Jobs[0].Name = "other-job"
	m2 := New(cfg2, nil)
	if err := m2.LoadState(tmp); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSaveState_BadPath(t *testing.T) {
	m := New(makeConfig(0), nil)
	err := m.SaveState(filepath.Join(t.TempDir(), "no", "such", "dir", "state.json"))
	if err == nil {
		t.Error("expected error for bad path, got nil")
	}
	_ = os.Remove("state.json")
}
