package monitor

import (
	"encoding/json"
	"os"
	"sync"
)

// TagStore manages arbitrary key-value tags associated with each job.
type TagStore struct {
	mu   sync.RWMutex
	tags map[string]map[string]string // job name -> tag key -> tag value
	path string
}

func NewTagStore(path string) *TagStore {
	s := &TagStore{
		tags: make(map[string]map[string]string),
		path: path,
	}
	_ = s.load()
	return s
}

func (s *TagStore) Set(job, key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tags[job]; !ok {
		s.tags[job] = make(map[string]string)
	}
	s.tags[job][key] = value
	_ = s.save()
}

func (s *TagStore) Get(job string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range s.tags[job] {
		result[k] = v
	}
	return result
}

func (s *TagStore) Delete(job, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.tags[job]; ok {
		delete(m, key)
		if len(m) == 0 {
			delete(s.tags, job)
		}
	}
	_ = s.save()
}

func (s *TagStore) save() error {
	data, err := json.Marshal(s.tags)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *TagStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.tags)
}
