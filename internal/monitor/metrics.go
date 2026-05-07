package monitor

import (
	"sync"
	"time"
)

// JobMetrics holds aggregated statistics for a single job.
type JobMetrics struct {
	JobName        string        `json:"job_name"`
	TotalHeartbeats int          `json:"total_heartbeats"`
	TotalMisses    int           `json:"total_misses"`
	LastHeartbeat  *time.Time    `json:"last_heartbeat,omitempty"`
	LastMissed     *time.Time    `json:"last_missed,omitempty"`
	UptimePct      float64       `json:"uptime_pct"`
	AvgLatency     time.Duration `json:"avg_latency_ns"`
}

// MetricsStore accumulates per-job metrics in memory.
type MetricsStore struct {
	mu      sync.RWMutex
	records map[string]*metricsRecord
}

type metricsRecord struct {
	totalHeartbeats int
	totalMisses     int
	lastHeartbeat   *time.Time
	lastMissed      *time.Time
	latencySum      time.Duration
}

// NewMetricsStore creates an empty MetricsStore.
func NewMetricsStore() *MetricsStore {
	return &MetricsStore{records: make(map[string]*metricsRecord)}
}

func (m *MetricsStore) ensure(job string) *metricsRecord {
	if _, ok := m.records[job]; !ok {
		m.records[job] = &metricsRecord{}
	}
	return m.records[job]
}

// RecordHeartbeat registers a successful heartbeat and its latency.
func (m *MetricsStore) RecordHeartbeat(job string, at time.Time, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.ensure(job)
	r.totalHeartbeats++
	r.lastHeartbeat = &at
	if latency > 0 {
		r.latencySum += latency
	}
}

// RecordMiss registers a missed execution for a job.
func (m *MetricsStore) RecordMiss(job string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := m.ensure(job)
	r.totalMisses++
	r.lastMissed = &at
}

// Get returns a snapshot of metrics for the given job.
func (m *MetricsStore) Get(job string) (JobMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.records[job]
	if !ok {
		return JobMetrics{}, false
	}
	total := r.totalHeartbeats + r.totalMisses
	var uptime float64
	if total > 0 {
		uptime = float64(r.totalHeartbeats) / float64(total) * 100.0
	}
	var avg time.Duration
	if r.totalHeartbeats > 0 {
		avg = r.latencySum / time.Duration(r.totalHeartbeats)
	}
	return JobMetrics{
		JobName:         job,
		TotalHeartbeats: r.totalHeartbeats,
		TotalMisses:     r.totalMisses,
		LastHeartbeat:   r.lastHeartbeat,
		LastMissed:      r.lastMissed,
		UptimePct:       uptime,
		AvgLatency:      avg,
	}, true
}

// All returns metrics for every tracked job.
func (m *MetricsStore) All() []JobMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]JobMetrics, 0, len(m.records))
	for job := range m.records {
		m.mu.RUnlock()
		v, _ := m.Get(job)
		m.mu.RLock()
		out = append(out, v)
	}
	return out
}
