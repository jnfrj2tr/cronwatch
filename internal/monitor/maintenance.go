package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// MaintenanceWindow represents a scheduled maintenance period for a job.
type MaintenanceWindow struct {
	JobName   string    `json:"job_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason,omitempty"`
}

// MaintenanceStore persists and queries maintenance windows.
type MaintenanceStore struct {
	mu       sync.RWMutex
	path     string
	windows  map[string][]MaintenanceWindow
	knownJobs map[string]struct{}
}

// NewMaintenanceStore creates a new MaintenanceStore for the given jobs.
func NewMaintenanceStore(path string, jobs []string) (*MaintenanceStore, error) {
	known := make(map[string]struct{}, len(jobs))
	for _, j := range jobs {
		known[j] = struct{}{}
	}
	s := &MaintenanceStore{path: path, windows: make(map[string][]MaintenanceWindow), knownJobs: known}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Set adds or replaces a maintenance window for a job.
func (s *MaintenanceStore) Set(w MaintenanceWindow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.knownJobs[w.JobName]; !ok {
		return ErrUnknownJob
	}
	s.windows[w.JobName] = append(s.windows[w.JobName], w)
	return s.save()
}

// IsUnderMaintenance returns true if the job has an active maintenance window at t.
func (s *MaintenanceStore) IsUnderMaintenance(job string, t time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.windows[job] {
		if !t.Before(w.StartTime) && t.Before(w.EndTime) {
			return true
		}
	}
	return false
}

// Get returns all maintenance windows for a job.
func (s *MaintenanceStore) Get(job string) ([]MaintenanceWindow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.knownJobs[job]; !ok {
		return nil, ErrUnknownJob
	}
	return append([]MaintenanceWindow(nil), s.windows[job]...), nil
}

// Delete removes all maintenance windows for a job.
func (s *MaintenanceStore) Delete(job string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.knownJobs[job]; !ok {
		return ErrUnknownJob
	}
	delete(s.windows, job)
	return s.save()
}

func (s *MaintenanceStore) save() error {
	data, err := json.Marshal(s.windows)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *MaintenanceStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var raw map[string][]MaintenanceWindow
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for job, ws := range raw {
		if _, ok := s.knownJobs[job]; ok {
			s.windows[job] = ws
		}
	}
	return nil
}
