package monitor

import (
	"encoding/json"
	"os"
	"sync"
)

// Annotation holds a user-defined note attached to a job.
type Annotation struct {
	JobName string `json:"job_name"`
	Note    string `json:"note"`
	Author  string `json:"author"`
}

// AnnotationStore persists per-job annotations.
type AnnotationStore struct {
	mu       sync.RWMutex
	path     string
	records  map[string]Annotation
}

// NewAnnotationStore creates a store backed by the given file path.
func NewAnnotationStore(path string) (*AnnotationStore, error) {
	s := &AnnotationStore{path: path, records: make(map[string]Annotation)}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Set stores an annotation for a job, overwriting any existing one.
func (s *AnnotationStore) Set(jobName, note, author string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[jobName] = Annotation{JobName: jobName, Note: note, Author: author}
	return s.save()
}

// Get returns the annotation for a job and whether it exists.
func (s *AnnotationStore) Get(jobName string) (Annotation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.records[jobName]
	return a, ok
}

// Delete removes the annotation for a job.
func (s *AnnotationStore) Delete(jobName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, jobName)
	return s.save()
}

func (s *AnnotationStore) save() error {
	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *AnnotationStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.records)
}
