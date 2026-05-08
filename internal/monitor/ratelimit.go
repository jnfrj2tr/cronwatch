package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// AlertRateLimit tracks when alerts were last sent per job to prevent alert storms.
type AlertRateLimit struct {
	mu       sync.Mutex
	path     string
	lastSent map[string]time.Time
	cooldown time.Duration
}

type rateLimitState struct {
	LastSent map[string]time.Time `json:"last_sent"`
}

// NewAlertRateLimit creates a new rate limiter with the given cooldown duration and persistence path.
func NewAlertRateLimit(path string, cooldown time.Duration) *AlertRateLimit {
	rl := &AlertRateLimit{
		path:     path,
		lastSent: make(map[string]time.Time),
		cooldown: cooldown,
	}
	rl.load()
	return rl
}

// Allow returns true if an alert for the given job is allowed (cooldown elapsed or never sent).
func (rl *AlertRateLimit) Allow(job string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	last, ok := rl.lastSent[job]
	if !ok {
		return true
	}
	return time.Since(last) >= rl.cooldown
}

// Record marks that an alert was sent for the given job right now.
func (rl *AlertRateLimit) Record(job string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastSent[job] = time.Now()
	return rl.save()
}

// Reset clears the rate limit record for a specific job.
func (rl *AlertRateLimit) Reset(job string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.lastSent, job)
	return rl.save()
}

func (rl *AlertRateLimit) save() error {
	state := rateLimitState{LastSent: rl.lastSent}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(rl.path, data, 0644)
}

func (rl *AlertRateLimit) load() {
	data, err := os.ReadFile(rl.path)
	if err != nil {
		return
	}
	var state rateLimitState
	if json.Unmarshal(data, &state) == nil && state.LastSent != nil {
		rl.lastSent = state.LastSent
	}
}
