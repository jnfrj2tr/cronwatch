package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempRateLimitPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "ratelimit.json")
}

func TestRateLimit_AllowFirstTime(t *testing.T) {
	rl := NewAlertRateLimit(tempRateLimitPath(t), 5*time.Minute)
	if !rl.Allow("job1") {
		t.Error("expected Allow to return true for unseen job")
	}
}

func TestRateLimit_BlockedWithinCooldown(t *testing.T) {
	rl := NewAlertRateLimit(tempRateLimitPath(t), 5*time.Minute)
	_ = rl.Record("job1")
	if rl.Allow("job1") {
		t.Error("expected Allow to return false within cooldown")
	}
}

func TestRateLimit_AllowedAfterCooldown(t *testing.T) {
	rl := NewAlertRateLimit(tempRateLimitPath(t), 1*time.Millisecond)
	_ = rl.Record("job1")
	time.Sleep(5 * time.Millisecond)
	if !rl.Allow("job1") {
		t.Error("expected Allow to return true after cooldown elapsed")
	}
}

func TestRateLimit_Reset(t *testing.T) {
	rl := NewAlertRateLimit(tempRateLimitPath(t), 5*time.Minute)
	_ = rl.Record("job1")
	_ = rl.Reset("job1")
	if !rl.Allow("job1") {
		t.Error("expected Allow to return true after reset")
	}
}

func TestRateLimit_PersistsAcrossReload(t *testing.T) {
	path := tempRateLimitPath(t)
	rl := NewAlertRateLimit(path, 5*time.Minute)
	_ = rl.Record("job1")

	rl2 := NewAlertRateLimit(path, 5*time.Minute)
	if rl2.Allow("job1") {
		t.Error("expected reloaded store to still block job1")
	}
}

func TestRateLimit_SaveBadPath(t *testing.T) {
	rl := NewAlertRateLimit("/nonexistent/dir/ratelimit.json", 5*time.Minute)
	err := rl.Record("job1")
	if err == nil {
		t.Error("expected error saving to bad path")
	}
}

func TestRateLimit_LoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	rl := NewAlertRateLimit(path, time.Minute)
	if !rl.Allow("anything") {
		t.Error("expected Allow true when no state file exists")
	}
}

func TestRateLimit_MultipleJobs(t *testing.T) {
	path := tempRateLimitPath(t)
	rl := NewAlertRateLimit(path, 5*time.Minute)
	_ = rl.Record("jobA")
	if !rl.Allow("jobB") {
		t.Error("expected jobB to be allowed when only jobA was recorded")
	}
	if rl.Allow("jobA") {
		t.Error("expected jobA to be blocked")
	}
	_ = os.Remove(path)
}
