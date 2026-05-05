package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	path := writeTempConfig(t, `
check_interval: 30s
log_file: /var/log/cronwatch.log
alerts:
  email: ops@example.com
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout: 10m
    command: /usr/local/bin/backup.sh
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Errorf("check_interval: got %v, want 30s", cfg.CheckInterval)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("job name: got %q, want %q", cfg.Jobs[0].Name, "backup")
	}
}

func TestLoad_DefaultCheckInterval(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: cleanup
    schedule: "@daily"
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != time.Minute {
		t.Errorf("default check_interval: got %v, want 1m", cfg.CheckInterval)
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `check_interval: 1m\n`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for config with no jobs")
	}
}

func TestLoad_DuplicateJobName(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: sync
    schedule: "@hourly"
  - name: sync
    schedule: "@daily"
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate job name")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
