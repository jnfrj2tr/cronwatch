package monitor

import (
	"testing"
	"time"
)

func TestMetrics_RecordHeartbeat(t *testing.T) {
	store := NewMetricsStore()
	now := time.Now()
	store.RecordHeartbeat("backup", now, 200*time.Millisecond)
	store.RecordHeartbeat("backup", now.Add(time.Minute), 400*time.Millisecond)

	m, ok := store.Get("backup")
	if !ok {
		t.Fatal("expected metrics for 'backup'")
	}
	if m.TotalHeartbeats != 2 {
		t.Errorf("expected 2 heartbeats, got %d", m.TotalHeartbeats)
	}
	if m.TotalMisses != 0 {
		t.Errorf("expected 0 misses, got %d", m.TotalMisses)
	}
	if m.UptimePct != 100.0 {
		t.Errorf("expected 100%% uptime, got %.2f", m.UptimePct)
	}
	if m.AvgLatency != 300*time.Millisecond {
		t.Errorf("expected avg latency 300ms, got %v", m.AvgLatency)
	}
}

func TestMetrics_RecordMiss(t *testing.T) {
	store := NewMetricsStore()
	now := time.Now()
	store.RecordHeartbeat("sync", now, 0)
	store.RecordMiss("sync", now.Add(time.Minute))
	store.RecordMiss("sync", now.Add(2*time.Minute))

	m, ok := store.Get("sync")
	if !ok {
		t.Fatal("expected metrics for 'sync'")
	}
	if m.TotalMisses != 2 {
		t.Errorf("expected 2 misses, got %d", m.TotalMisses)
	}
	expected := 100.0 / 3.0
	if m.UptimePct < expected-0.01 || m.UptimePct > expected+0.01 {
		t.Errorf("expected uptime ~%.2f%%, got %.2f%%", expected, m.UptimePct)
	}
}

func TestMetrics_GetUnknownJob(t *testing.T) {
	store := NewMetricsStore()
	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestMetrics_All(t *testing.T) {
	store := NewMetricsStore()
	now := time.Now()
	store.RecordHeartbeat("jobA", now, 0)
	store.RecordMiss("jobB", now)

	all := store.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestMetrics_UptimeAllMisses(t *testing.T) {
	store := NewMetricsStore()
	now := time.Now()
	store.RecordMiss("nightly", now)
	store.RecordMiss("nightly", now.Add(time.Hour))

	m, _ := store.Get("nightly")
	if m.UptimePct != 0.0 {
		t.Errorf("expected 0%% uptime, got %.2f", m.UptimePct)
	}
}
