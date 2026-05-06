package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempSilencePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "silences.json")
}

func TestSilenceStore_IsSilenced_Active(t *testing.T) {
	store := NewSilenceStore(tempSilencePath(t))
	now := time.Now()
	_ = store.Add(Silence{
		JobName:   "backup",
		StartTime: now.Add(-10 * time.Minute),
		EndTime:   now.Add(10 * time.Minute),
		Reason:    "maintenance",
	})
	if !store.IsSilenced("backup", now) {
		t.Error("expected job to be silenced")
	}
}

func TestSilenceStore_IsSilenced_Expired(t *testing.T) {
	store := NewSilenceStore(tempSilencePath(t))
	now := time.Now()
	_ = store.Add(Silence{
		JobName:   "backup",
		StartTime: now.Add(-20 * time.Minute),
		EndTime:   now.Add(-5 * time.Minute),
	})
	if store.IsSilenced("backup", now) {
		t.Error("expected expired silence to not match")
	}
}

func TestSilenceStore_IsSilenced_UnknownJob(t *testing.T) {
	store := NewSilenceStore(tempSilencePath(t))
	now := time.Now()
	_ = store.Add(Silence{
		JobName:   "other-job",
		StartTime: now.Add(-5 * time.Minute),
		EndTime:   now.Add(5 * time.Minute),
	})
	if store.IsSilenced("backup", now) {
		t.Error("expected different job not to be silenced")
	}
}

func TestSilenceStore_Purge(t *testing.T) {
	path := tempSilencePath(t)
	store := NewSilenceStore(path)
	now := time.Now()
	_ = store.Add(Silence{JobName: "a", StartTime: now.Add(-30 * time.Minute), EndTime: now.Add(-1 * time.Minute)})
	_ = store.Add(Silence{JobName: "b", StartTime: now.Add(-5 * time.Minute), EndTime: now.Add(10 * time.Minute)})

	if err := store.Purge(now); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	if store.IsSilenced("a", now) {
		t.Error("expected expired silence to be purged")
	}
	if !store.IsSilenced("b", now) {
		t.Error("expected active silence to remain after purge")
	}
}

func TestSilenceStore_PersistsAcrossReload(t *testing.T) {
	path := tempSilencePath(t)
	now := time.Now()

	store1 := NewSilenceStore(path)
	_ = store1.Add(Silence{
		JobName:   "deploy",
		StartTime: now.Add(-1 * time.Minute),
		EndTime:   now.Add(60 * time.Minute),
	})

	store2 := NewSilenceStore(path)
	if !store2.IsSilenced("deploy", now) {
		t.Error("expected silence to be loaded from disk")
	}
}

func TestSilenceStore_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "silences.json")
	store := NewSilenceStore(path)
	if store.IsSilenced("any", time.Now()) {
		t.Error("expected empty store for missing file")
	}
	if err := store.Add(Silence{JobName: "x", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour)}); err == nil {
		_ = os.Remove(path)
	}
}
