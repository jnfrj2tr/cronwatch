package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistory_AddAndRecent(t *testing.T) {
	h := NewHistory("", 3)
	for i := 0; i < 5; i++ {
		h.Add(RunRecord{
			JobName:   "job",
			Timestamp: time.Now(),
			Status:    "ok",
		})
	}
	if len(h.Records) != 3 {
		t.Fatalf("expected 3 records after eviction, got %d", len(h.Records))
	}
}

func TestHistory_RecentLimitedByN(t *testing.T) {
	h := NewHistory("", 10)
	for i := 0; i < 5; i++ {
		h.Add(RunRecord{JobName: "job", Timestamp: time.Now(), Status: "ok"})
	}
	recent := h.Recent(2)
	if len(recent) != 2 {
		t.Fatalf("expected 2 recent records, got %d", len(recent))
	}
}

func TestHistory_RecentAllWhenFewerThanN(t *testing.T) {
	h := NewHistory("", 10)
	h.Add(RunRecord{JobName: "job", Timestamp: time.Now(), Status: "missed"})
	recent := h.Recent(5)
	if len(recent) != 1 {
		t.Fatalf("expected 1 recent record, got %d", len(recent))
	}
}

func TestHistory_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h := NewHistory(path, 10)
	h.Add(RunRecord{JobName: "backup", Timestamp: time.Now().UTC().Truncate(time.Second), Status: "ok"})
	h.Add(RunRecord{JobName: "cleanup", Timestamp: time.Now().UTC().Truncate(time.Second), Status: "missed"})

	if err := h.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	h2 := NewHistory(path, 10)
	if err := h2.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(h2.Records) != 2 {
		t.Fatalf("expected 2 records after load, got %d", len(h2.Records))
	}
	if h2.Records[0].JobName != "backup" {
		t.Errorf("expected first record job 'backup', got %q", h2.Records[0].JobName)
	}
}

func TestHistory_LoadMissingFile(t *testing.T) {
	h := NewHistory("/nonexistent/path/history.json", 10)
	if err := h.Load(); err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(h.Records) != 0 {
		t.Errorf("expected empty records, got %d", len(h.Records))
	}
}

func TestHistory_SaveBadPath(t *testing.T) {
	h := NewHistory("/nonexistent/dir/history.json", 10)
	h.Add(RunRecord{JobName: "job", Timestamp: time.Now(), Status: "ok"})
	if err := h.Save(); err == nil {
		t.Error("expected error saving to bad path, got nil")
	}
	_ = os.Remove("/nonexistent/dir/history.json")
}
