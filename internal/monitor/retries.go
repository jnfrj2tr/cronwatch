package monitor

import (
	"encoding/json"
	"os"
	"sync"
)

// RetryPolicy defines how many times a missed job alert should be retried
// before escalating or suppressing further notifications.
type RetryPolicy struct {
	MaxRetries int `yaml:"max_retries" json:"max_retries"`
	BackoffSec int `yaml:"backoff_sec" json:"backoff_sec"`
}

type retryEntry struct {
	Count int `json:"count"`
}

// RetryStore tracks per-job alert retry counts.
type RetryStore struct {
	mu      sync.Mutex
	entries map[string]*retryEntry
	path    string
}

// NewRetryStore creates a RetryStore backed by the given file path.
func NewRetryStore(path string, knownJobs []string) (*RetryStore, error) {
	s := &RetryStore{
		entries: make(map[string]*retryEntry),
		path:    path,
	}
	for _, j := range knownJobs {
		s.entries[j] = &retryEntry{}
	}
	if err := s.load(knownJobs); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *RetryStore) load(knownJobs []string) error {
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var raw map[string]*retryEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	known := make(map[string]bool, len(knownJobs))
	for _, j := range knownJobs {
		known[j] = true
	}
	for k, v := range raw {
		if known[k] {
			s.entries[k] = v
		}
	}
	return nil
}

func (s *RetryStore) save() error {
	data, err := json.Marshal(s.entries)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

// Increment increases the retry count for a job and returns the new count.
func (s *RetryStore) Increment(job string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return 0, ErrUnknownJob
	}
	e.Count++
	return e.Count, s.save()
}

// Reset clears the retry count for a job (e.g. after a successful heartbeat).
func (s *RetryStore) Reset(job string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return ErrUnknownJob
	}
	e.Count = 0
	return s.save()
}

// Get returns the current retry count for a job.
func (s *RetryStore) Get(job string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return 0, ErrUnknownJob
	}
	return e.Count, nil
}
