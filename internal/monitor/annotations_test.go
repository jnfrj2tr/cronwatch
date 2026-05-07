package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

func tempAnnotationPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "annotations.json")
}

func TestAnnotationStore_SetAndGet(t *testing.T) {
	store, err := NewAnnotationStore(tempAnnotationPath(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := store.Set("backup", "runs nightly", "alice"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	a, ok := store.Get("backup")
	if !ok {
		t.Fatal("expected annotation to exist")
	}
	if a.Note != "runs nightly" || a.Author != "alice" {
		t.Errorf("unexpected annotation: %+v", a)
	}
}

func TestAnnotationStore_GetUnknownJob(t *testing.T) {
	store, _ := NewAnnotationStore(tempAnnotationPath(t))
	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("expected no annotation for unknown job")
	}
}

func TestAnnotationStore_Delete(t *testing.T) {
	store, _ := NewAnnotationStore(tempAnnotationPath(t))
	_ = store.Set("cleanup", "weekly", "bob")
	if err := store.Delete("cleanup"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, ok := store.Get("cleanup")
	if ok {
		t.Error("expected annotation to be deleted")
	}
}

func TestAnnotationStore_PersistsAcrossReload(t *testing.T) {
	path := tempAnnotationPath(t)
	store, _ := NewAnnotationStore(path)
	_ = store.Set("sync", "important", "carol")

	reloaded, err := NewAnnotationStore(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	a, ok := reloaded.Get("sync")
	if !ok || a.Note != "important" {
		t.Errorf("expected persisted annotation, got %+v", a)
	}
}

func TestAnnotationStore_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	store, err := NewAnnotationStore(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestAnnotationStore_BadPath(t *testing.T) {
	store, _ := NewAnnotationStore("/nonexistent/dir/annotations.json")
	if store != nil {
		// store creation may succeed but save should fail
		err := store.Set("job", "note", "user")
		if err == nil {
			t.Error("expected error writing to bad path")
		}
	}
	_ = os.Remove("/nonexistent/dir/annotations.json")
}
