package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// RunRecord represents a single recorded execution of a cron job.
type RunRecord struct {
	JobName   string    `json:"job_name"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "ok" or "missed"
}

// History maintains a bounded in-memory log of recent job run records
// and can persist them to disk.
type History struct {
	mu      sync.Mutex
	Records []RunRecord `json:"records"`
	MaxSize int         `json:"-"`
	path    string
}

// NewHistory creates a History with the given max size and file path.
func NewHistory(path string, maxSize int) *History {
	return &History{
		path:    path,
		MaxSize: maxSize,
	}
}

// Add appends a RunRecord, evicting the oldest entry if capacity is exceeded.
func (h *History) Add(r RunRecord) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Records = append(h.Records, r)
	if h.MaxSize > 0 && len(h.Records) > h.MaxSize {
		h.Records = h.Records[len(h.Records)-h.MaxSize:]
	}
}

// Recent returns up to n most recent records across all jobs.
func (h *History) Recent(n int) []RunRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	if n >= len(h.Records) {
		out := make([]RunRecord, len(h.Records))
		copy(out, h.Records)
		return out
	}
	out := make([]RunRecord, n)
	copy(out, h.Records[len(h.Records)-n:])
	return out
}

// Save persists the history to the configured file path as JSON.
func (h *History) Save() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0644)
}

// Load reads history from disk, ignoring missing-file errors.
func (h *History) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	data, err := os.ReadFile(h.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, h)
}
