package monitor

import (
	"log"
	"sync"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/schedule"
)

// JobStatus holds the last known state of a monitored job.
type JobStatus struct {
	Name        string
	LastSeen    time.Time
	Missed      bool
	Consecutive int
}

// AlertFunc is called when a job is detected as missed or failed.
type AlertFunc func(job config.Job, status JobStatus)

// Monitor checks cron jobs at a configured interval.
type Monitor struct {
	cfg      *config.Config
	status   map[string]*JobStatus
	mu       sync.Mutex
	alertFn  AlertFunc
	stopCh   chan struct{}
}

// New creates a new Monitor with the given config and alert function.
func New(cfg *config.Config, alertFn AlertFunc) *Monitor {
	status := make(map[string]*JobStatus, len(cfg.Jobs))
	for _, j := range cfg.Jobs {
		status[j.Name] = &JobStatus{Name: j.Name}
	}
	return &Monitor{
		cfg:     cfg,
		status:  status,
		alertFn: alertFn,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the monitoring loop.
func (m *Monitor) Start() {
	ticker := time.NewTicker(time.Duration(m.cfg.CheckInterval) * time.Second)
	defer ticker.Stop()
	log.Printf("monitor: starting, check interval %ds", m.cfg.CheckInterval)
	for {
		select {
		case <-ticker.C:
			m.check(time.Now())
		case <-m.stopCh:
			log.Println("monitor: stopped")
			return
		}
	}
}

// Stop signals the monitoring loop to exit.
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// RecordHeartbeat marks a job as seen at the given time.
func (m *Monitor) RecordHeartbeat(name string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.status[name]; ok {
		s.LastSeen = at
		s.Missed = false
		s.Consecutive = 0
	}
}

// Status returns a copy of the current JobStatus for the named job.
// The second return value is false if no job with that name is tracked.
func (m *Monitor) Status(name string) (JobStatus, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.status[name]
	if !ok {
		return JobStatus{}, false
	}
	return *s, true
}

// check evaluates each job against its expected schedule.
func (m *Monitor) check(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, job := range m.cfg.Jobs {
		s := m.status[job.Name]
		last, err := schedule.LastExpected(job.Schedule, now)
		if err != nil {
			log.Printf("monitor: invalid schedule for %q: %v", job.Name, err)
			continue
		}
		grace := time.Duration(job.GraceSeconds) * time.Second
		if s.LastSeen.Before(last.Add(-grace)) {
			s.Missed = true
			s.Consecutive++
			log.Printf("monitor: missed job %q (consecutive=%d)", job.Name, s.Consecutive)
			if m.alertFn != nil {
				m.alertFn(job, *s)
			}
		}
	}
}
