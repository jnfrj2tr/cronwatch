package monitor

import (
	"encoding/json"
	"os"
	"time"
)

// persistedState is the on-disk representation of all job statuses.
type persistedState struct {
	UpdatedAt time.Time             `json:"updated_at"`
	Jobs      map[string]JobStatus  `json:"jobs"`
}

// SaveState writes the current job statuses to a JSON file.
func (m *Monitor) SaveState(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ps := persistedState{
		UpdatedAt: time.Now(),
		Jobs:      make(map[string]JobStatus, len(m.status)),
	}
	for k, v := range m.status {
		ps.Jobs[k] = *v
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(ps)
}

// LoadState reads persisted job statuses from a JSON file into the monitor.
// Unknown jobs in the file are silently ignored.
func (m *Monitor) LoadState(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // first run — no state yet
		}
		return err
	}
	defer f.Close()

	var ps persistedState
	if err := json.NewDecoder(f).Decode(&ps); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for name, saved := range ps.Jobs {
		if s, ok := m.status[name]; ok {
			*s = saved
		}
	}
	return nil
}
