package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Annotation struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type AnnotationStore struct {
	mu   sync.Mutex
	path string
	data map[string][]Annotation
}

func NewAnnotationStore(path string) *AnnotationStore {
	s := &AnnotationStore{path: path, data: make(map[string][]Annotation)}
	s.load()
	return s
}

func (s *AnnotationStore) Add(job, text string, at time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[job] = append(s.data[job], Annotation{Text: text, CreatedAt: at})
	s.save()
}

func (s *AnnotationStore) Get(job string) []Annotation {
	s.mu.Lock()
	defer s.mu.Unlock()
	if annotations, ok := s.data[job]; ok {
		copy := make([]Annotation, len(annotations))
		for i, a := range annotations {
			copy[i] = a
		}
		return copy
	}
	return []Annotation{}
}

func (s *AnnotationStore) Clear(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, job)
	s.save()
}

func (s *AnnotationStore) All() map[string][]Annotation {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[string][]Annotation, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

func (s *AnnotationStore) load() {
	f, err := os.Open(s.path)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewDecoder(f).Decode(&s.data)
}

func (s *AnnotationStore) save() {
	f, err := os.Create(s.path)
	if err != nil {
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(s.data)
}
