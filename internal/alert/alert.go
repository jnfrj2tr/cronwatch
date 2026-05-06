package alert

import (
	"fmt"
	"log"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert represents a single alert event.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	OccurredAt time.Time
}

// Notifier sends alerts via a specific channel.
type Notifier interface {
	Send(a Alert) error
}

// Manager dispatches alerts to one or more notifiers.
type Manager struct {
	notifiers []Notifier
}

// NewManager creates a Manager with the given notifiers.
func NewManager(notifiers ...Notifier) *Manager {
	return &Manager{notifiers: notifiers}
}

// Send dispatches an alert to all registered notifiers.
func (m *Manager) Send(jobName string, level Level, msg string) {
	a := Alert{
		JobName:    jobName,
		Level:      level,
		Message:    msg,
		OccurredAt: time.Now(),
	}
	for _, n := range m.notifiers {
		if err := n.Send(a); err != nil {
			log.Printf("alert: notifier error: %v", err)
		}
	}
}

// FormatMessage returns a human-readable alert string.
func FormatMessage(a Alert) string {
	return fmt.Sprintf("[%s] %s — %s (at %s)",
		a.Level, a.JobName, a.Message,
		a.OccurredAt.Format(time.RFC3339))
}
