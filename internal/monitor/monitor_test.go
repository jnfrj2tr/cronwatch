package monitor

import (
	"sync"
	"testing"
	"time"

	"github.com/cronwatch/internal/config"
)

func makeConfig(graceSec int) *config.Config {
	return &config.Config{
		CheckInterval: 60,
		Jobs: []config.Job{
			{
				Name:         "test-job",
				Schedule:     "* * * * *",
				GraceSeconds: graceSec,
			},
		},
	}
}

func TestRecordHeartbeat(t *testing.T) {
	m := New(makeConfig(30), nil)
	now := time.Now()
	m.RecordHeartbeat("test-job", now)
	m.mu.Lock()
	defer m.mu.Unlock()
	if got := m.status["test-job"].LastSeen; !got.Equal(now) {
		t.Errorf("expected LastSeen=%v, got %v", now, got)
	}
}

func TestCheck_MissedJob(t *testing.T) {
	var mu sync.Mutex
	var alerts []string
	alertFn := func(job config.Job, s JobStatus) {
		mu.Lock()
		alerts = append(alerts, job.Name)
		mu.Unlock()
	}
	m := New(makeConfig(0), alertFn)
	// LastSeen is zero — job has never run, should be missed
	m.check(time.Now())
	mu.Lock()
	defer mu.Unlock()
	if len(alerts) == 0 {
		t.Error("expected at least one alert for missed job")
	}
}

func TestCheck_JobWithinGrace(t *testing.T) {
	var alerts []string
	alertFn := func(job config.Job, s JobStatus) {
		alerts = append(alerts, job.Name)
	}
	m := New(makeConfig(3600), alertFn)
	// Record heartbeat just now — within any grace period
	m.RecordHeartbeat("test-job", time.Now())
	m.check(time.Now())
	if len(alerts) != 0 {
		t.Errorf("expected no alerts, got %d", len(alerts))
	}
}

func TestCheck_ConsecutiveMisses(t *testing.T) {
	m := New(makeConfig(0), nil)
	now := time.Now()
	m.check(now)
	m.check(now)
	m.mu.Lock()
	defer m.mu.Unlock()
	if got := m.status["test-job"].Consecutive; got != 2 {
		t.Errorf("expected Consecutive=2, got %d", got)
	}
}

func TestRecordHeartbeat_ResetsMissed(t *testing.T) {
	m := New(makeConfig(0), nil)
	m.check(time.Now())
	m.RecordHeartbeat("test-job", time.Now())
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status["test-job"].Missed {
		t.Error("expected Missed=false after heartbeat")
	}
	if m.status["test-job"].Consecutive != 0 {
		t.Error("expected Consecutive=0 after heartbeat")
	}
}
