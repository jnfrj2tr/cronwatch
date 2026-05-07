package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempRunLogPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "runlog.json")
}

func TestRunLog_RecordAndRecent(t *testing.T) {
	rl, err := NewRunLog(tempRunLogPath(t), 10)
	if err != nil {
		t.Fatalf("NewRunLog: %v", err)
	}
	now := time.Now()
	r := RunResult{JobName: "backup", StartedAt: now, FinishedAt: now.Add(2 * time.Second), Duration: 2 * time.Second, Success: true}
	if err := rl.Record(r); err != nil {
		t.Fatalf("Record: %v", err)
	}
	results := rl.Recent("backup", 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Error("expected success=true")
	}
}

func TestRunLog_MaxPerJobEnforced(t *testing.T) {
	rl, err := NewRunLog(tempRunLogPath(t), 3)
	if err != nil {
		t.Fatalf("NewRunLog: %v", err)
	}
	now := time.Now()
	for i := 0; i < 5; i++ {
		_ = rl.Record(RunResult{JobName: "job", StartedAt: now, FinishedAt: now, Success: i%2 == 0})
	}
	results := rl.Recent("job", 10)
	if len(results) != 3 {
		t.Fatalf("expected 3 results (capped), got %d", len(results))
	}
}

func TestRunLog_RecentUnknownJob(t *testing.T) {
	rl, _ := NewRunLog(tempRunLogPath(t), 10)
	results := rl.Recent("nonexistent", 5)
	if len(results) != 0 {
		t.Errorf("expected 0 results for unknown job, got %d", len(results))
	}
}

func TestRunLog_PersistsAcrossReload(t *testing.T) {
	path := tempRunLogPath(t)
	rl, _ := NewRunLog(path, 10)
	now := time.Now()
	_ = rl.Record(RunResult{JobName: "sync", StartedAt: now, FinishedAt: now, Success: false, ExitCode: 1, Message: "timeout"})

	rl2, err := NewRunLog(path, 10)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	results := rl2.Recent("sync", 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 result after reload, got %d", len(results))
	}
	if results[0].Message != "timeout" {
		t.Errorf("expected message 'timeout', got %q", results[0].Message)
	}
}

func TestRunLog_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	rl, err := NewRunLog(path, 10)
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if rl == nil {
		t.Fatal("expected non-nil RunLog")
	}
}

func TestRunLog_BadPath(t *testing.T) {
	_, err := NewRunLog("/nonexistent/dir/runlog.json", 10)
	if err == nil {
		t.Error("expected error for bad path")
	}
	// Ensure a file at that path doesn't accidentally exist
	if _, statErr := os.Stat("/nonexistent/dir/runlog.json"); statErr == nil {
		t.Skip("path unexpectedly exists")
	}
}
