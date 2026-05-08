package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

func tempRetryPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "retries.json")
}

func TestRetryStore_IncrementAndGet(t *testing.T) {
	s, err := NewRetryStore(tempRetryPath(t), []string{"backup"})
	if err != nil {
		t.Fatal(err)
	}
	count, err := s.Increment("backup")
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
	count, err = s.Increment("backup")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
	got, err := s.Get("backup")
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Errorf("expected 2, got %d", got)
	}
}

func TestRetryStore_Reset(t *testing.T) {
	s, err := NewRetryStore(tempRetryPath(t), []string{"sync"})
	if err != nil {
		t.Fatal(err)
	}
	s.Increment("sync")
	s.Increment("sync")
	if err := s.Reset("sync"); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get("sync")
	if got != 0 {
		t.Errorf("expected 0 after reset, got %d", got)
	}
}

func TestRetryStore_UnknownJob(t *testing.T) {
	s, err := NewRetryStore(tempRetryPath(t), []string{"known"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Increment("ghost")
	if err != ErrUnknownJob {
		t.Errorf("expected ErrUnknownJob, got %v", err)
	}
	_, err = s.Get("ghost")
	if err != ErrUnknownJob {
		t.Errorf("expected ErrUnknownJob, got %v", err)
	}
	if err := s.Reset("ghost"); err != ErrUnknownJob {
		t.Errorf("expected ErrUnknownJob, got %v", err)
	}
}

func TestRetryStore_PersistsAcrossReload(t *testing.T) {
	path := tempRetryPath(t)
	s1, err := NewRetryStore(path, []string{"report"})
	if err != nil {
		t.Fatal(err)
	}
	s1.Increment("report")
	s1.Increment("report")
	s1.Increment("report")

	s2, err := NewRetryStore(path, []string{"report"})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s2.Get("report")
	if got != 3 {
		t.Errorf("expected 3 after reload, got %d", got)
	}
}

func TestRetryStore_BadPath(t *testing.T) {
	_, err := NewRetryStore("/nonexistent/dir/retries.json", []string{"job"})
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func TestRetryStore_UnknownJobIgnoredOnLoad(t *testing.T) {
	path := tempRetryPath(t)
	// seed with a job that won't be in the next load's known list
	s1, _ := NewRetryStore(path, []string{"old"})
	s1.Increment("old")

	s2, err := NewRetryStore(path, []string{"new"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s2.Get("old")
	if err != ErrUnknownJob {
		t.Error("old job should not be present after reload with different known jobs")
	}
	_ = os.Remove(path)
}
