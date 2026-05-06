package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Silence represents a time window during which alerts for a job are suppressed.
type Silence struct {
	JobName   string    `json:"job_name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Reason    string    `json:"reason,omitempty"`
}

// SilenceStore manages active silences for jobs.
type SilenceStore struct {
	mu       sync.RWMutex
	silences []Silence
	path     string
}

// NewSilenceStore creates a new SilenceStore backed by the given file path.
func NewSilenceStore(path string) *SilenceStore {
	s := &SilenceStore{path: path}
	_ = s.load()
	return s
}

// Add inserts a new silence window for the given job.
func (s *SilenceStore) Add(sil Silence) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.silences = append(s.silences, sil)
	return s.save()
}

// IsSilenced reports whether the given job is silenced at the given time.
func (s *SilenceStore) IsSilenced(jobName string, at time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sil := range s.silences {
		if sil.JobName == jobName && !at.Before(sil.StartTime) && at.Before(sil.EndTime) {
			return true
		}
	}
	return false
}

// Purge removes silences that have already expired relative to now.
func (s *SilenceStore) Purge(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	active := s.silences[:0]
	for _, sil := range s.silences {
		if sil.EndTime.After(now) {
			active = append(active, sil)
		}
	}
	s.silences = active
	return s.save()
}

func (s *SilenceStore) save() error {
	data, err := json.Marshal(s.silences)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *SilenceStore) load() error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.silences)
}
