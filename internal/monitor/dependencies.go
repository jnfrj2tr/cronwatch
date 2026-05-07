package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// DependencyStore tracks job dependencies — jobs that must succeed before another runs.
type DependencyStore struct {
	mu   sync.RWMutex
	path string
	deps map[string][]string // job -> list of jobs it depends on
	known map[string]bool
}

func NewDependencyStore(path string, knownJobs []string) (*DependencyStore, error) {
	ds := &DependencyStore{
		path:  path,
		deps:  make(map[string][]string),
		known: make(map[string]bool),
	}
	for _, j := range knownJobs {
		ds.known[j] = true
	}
	if err := ds.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return ds, nil
}

func (ds *DependencyStore) Set(job string, deps []string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if !ds.known[job] {
		return fmt.Errorf("unknown job: %s", job)
	}
	for _, d := range deps {
		if !ds.known[d] {
			return fmt.Errorf("unknown dependency job: %s", d)
		}
	}
	ds.deps[job] = deps
	return ds.save()
}

func (ds *DependencyStore) Get(job string) ([]string, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if !ds.known[job] {
		return nil, false
	}
	return ds.deps[job], true
}

func (ds *DependencyStore) Delete(job string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.deps, job)
	return ds.save()
}

func (ds *DependencyStore) All() map[string][]string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	out := make(map[string][]string, len(ds.deps))
	for k, v := range ds.deps {
		out[k] = v
	}
	return out
}

func (ds *DependencyStore) save() error {
	data, err := json.Marshal(ds.deps)
	if err != nil {
		return err
	}
	return os.WriteFile(ds.path, data, 0644)
}

func (ds *DependencyStore) load() error {
	data, err := os.ReadFile(ds.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &ds.deps)
}
