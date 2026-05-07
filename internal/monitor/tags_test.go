package monitor

import (
	"os"
	"path/filepath"
	"testing"
)

func tempTagPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "tags.json")
}

func TestTagStore_SetAndGet(t *testing.T) {
	s := NewTagStore(tempTagPath(t))
	s.Set("backup", "env", "prod")
	s.Set("backup", "team", "infra")

	tags := s.Get("backup")
	if tags["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", tags["env"])
	}
	if tags["team"] != "infra" {
		t.Errorf("expected team=infra, got %s", tags["team"])
	}
}

func TestTagStore_GetUnknownJob(t *testing.T) {
	s := NewTagStore(tempTagPath(t))
	tags := s.Get("nonexistent")
	if len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", tags)
	}
}

func TestTagStore_Delete(t *testing.T) {
	s := NewTagStore(tempTagPath(t))
	s.Set("job1", "key", "val")
	s.Delete("job1", "key")
	tags := s.Get("job1")
	if _, ok := tags["key"]; ok {
		t.Error("expected key to be deleted")
	}
}

func TestTagStore_PersistsAcrossReload(t *testing.T) {
	path := tempTagPath(t)
	s1 := NewTagStore(path)
	s1.Set("myjob", "region", "us-east-1")

	s2 := NewTagStore(path)
	tags := s2.Get("myjob")
	if tags["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1 after reload, got %s", tags["region"])
	}
}

func TestTagStore_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notexist.json")
	s := NewTagStore(path)
	tags := s.Get("any")
	if len(tags) != 0 {
		t.Errorf("expected empty on missing file, got %v", tags)
	}
}

func TestTagStore_SaveBadPath(t *testing.T) {
	s := NewTagStore("/nonexistent/dir/tags.json")
	s.Set("job", "k", "v") // should not panic
	_ = os.Remove("/nonexistent/dir/tags.json")
}
