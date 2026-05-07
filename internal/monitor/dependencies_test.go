package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

func tempDepPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "deps.json")
}

func TestDependencyStore_SetAndGet(t *testing.T) {
	ds, err := NewDependencyStore(tempDepPath(t), []string{"jobA", "jobB", "jobC"})
	if err != nil {
		t.Fatal(err)
	}
	if err := ds.Set("jobC", []string{"jobA", "jobB"}); err != nil {
		t.Fatal(err)
	}
	deps, ok := ds.Get("jobC")
	if !ok {
		t.Fatal("expected deps for jobC")
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}
}

func TestDependencyStore_GetUnknownJob(t *testing.T) {
	ds, _ := NewDependencyStore(tempDepPath(t), []string{"jobA"})
	_, ok := ds.Get("ghost")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestDependencyStore_SetUnknownJob(t *testing.T) {
	ds, _ := NewDependencyStore(tempDepPath(t), []string{"jobA"})
	if err := ds.Set("ghost", []string{"jobA"}); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestDependencyStore_SetUnknownDep(t *testing.T) {
	ds, _ := NewDependencyStore(tempDepPath(t), []string{"jobA", "jobB"})
	if err := ds.Set("jobA", []string{"ghost"}); err == nil {
		t.Fatal("expected error for unknown dependency")
	}
}

func TestDependencyStore_Delete(t *testing.T) {
	ds, _ := NewDependencyStore(tempDepPath(t), []string{"jobA", "jobB"})
	_ = ds.Set("jobA", []string{"jobB"})
	_ = ds.Delete("jobA")
	deps, ok := ds.Get("jobA")
	if !ok {
		t.Fatal("job should still be known after delete")
	}
	if len(deps) != 0 {
		t.Fatal("expected empty deps after delete")
	}
}

func TestDependencyStore_PersistsAcrossReload(t *testing.T) {
	path := tempDepPath(t)
	ds, _ := NewDependencyStore(path, []string{"jobA", "jobB"})
	_ = ds.Set("jobA", []string{"jobB"})

	ds2, err := NewDependencyStore(path, []string{"jobA", "jobB"})
	if err != nil {
		t.Fatal(err)
	}
	deps, ok := ds2.Get("jobA")
	if !ok || len(deps) != 1 || deps[0] != "jobB" {
		t.Fatalf("expected persisted deps, got %v", deps)
	}
}

func TestDependencyStore_LoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	ds, err := NewDependencyStore(path, []string{"jobA"})
	if err != nil {
		t.Fatal("expected no error on missing file")
	}
	if ds == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestDependencyStore_All(t *testing.T) {
	ds, _ := NewDependencyStore(tempDepPath(t), []string{"jobA", "jobB", "jobC"})
	_ = ds.Set("jobB", []string{"jobA"})
	_ = ds.Set("jobC", []string{"jobA", "jobB"})
	all := ds.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestDependencyStore_SaveBadPath(t *testing.T) {
	ds, _ := NewDependencyStore(tempDepPath(t), []string{"jobA", "jobB"})
	ds.path = "/nonexistent/dir/deps.json"
	if err := ds.Set("jobA", []string{"jobB"}); err == nil {
		t.Fatal("expected error for bad path")
	}
	_ = os.Remove(ds.path)
}
