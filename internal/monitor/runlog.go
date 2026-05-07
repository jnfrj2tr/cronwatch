package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// RunResult represents the outcome of a single cron job execution.
type RunResult struct {
	JobName   string        `json:"job_name"`
	StartedAt time.Time     `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	Duration  time.Duration `json:"duration_ms"`
	Success   bool          `json:"success"`
	ExitCode  int           `json:"exit_code"`
	Message   string        `json:"message,omitempty"`
}

// RunLog stores recent run results per job.
type RunLog struct {
	mu      sync.RWMutex
	path    string
	entries map[string][]RunResult
	maxPer  int
}

// NewRunLog creates a RunLog backed by the given file path.
func NewRunLog(path string, maxPerJob int) (*RunLog, error) {
	rl := &RunLog{
		path:    path,
		entries: make(map[string][]RunResult),
		maxPer:  maxPerJob,
	}
	if err := rl.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return rl, nil
}

// Record appends a run result for the given job.
func (rl *RunLog) Record(r RunResult) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	list := rl.entries[r.JobName]
	list = append(list, r)
	if len(list) > rl.maxPer {
		list = list[len(list)-rl.maxPer:]
	}
	rl.entries[r.JobName] = list
	return rl.save()
}

// Recent returns the last n run results for a job.
func (rl *RunLog) Recent(jobName string, n int) []RunResult {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	list := rl.entries[jobName]
	if n <= 0 || n >= len(list) {
		out := make([]RunResult, len(list))
		copy(out, list)
		return out
	}
	out := make([]RunResult, n)
	copy(out, list[len(list)-n:])
	return out
}

func (rl *RunLog) save() error {
	data, err := json.MarshalIndent(rl.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(rl.path, data, 0644)
}

func (rl *RunLog) load() error {
	data, err := os.ReadFile(rl.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &rl.entries)
}
