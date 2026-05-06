package alert_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
)

// fakeNotifier records alerts and optionally returns an error.
type fakeNotifier struct {
	received []alert.Alert
	errOnSend error
}

func (f *fakeNotifier) Send(a alert.Alert) error {
	f.received = append(f.received, a)
	return f.errOnSend
}

func TestManager_SendDispatchesToAllNotifiers(t *testing.T) {
	n1 := &fakeNotifier{}
	n2 := &fakeNotifier{}
	m := alert.NewManager(n1, n2)

	m.Send("backup", alert.LevelError, "missed run")

	if len(n1.received) != 1 || len(n2.received) != 1 {
		t.Fatalf("expected 1 alert each, got n1=%d n2=%d", len(n1.received), len(n2.received))
	}
	if n1.received[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", n1.received[0].JobName)
	}
}

func TestManager_ContinuesOnNotifierError(t *testing.T) {
	n1 := &fakeNotifier{errOnSend: errors.New("fail")}
	n2 := &fakeNotifier{}
	m := alert.NewManager(n1, n2)

	m.Send("job", alert.LevelWarn, "late")

	// n2 should still receive the alert despite n1 failing
	if len(n2.received) != 1 {
		t.Errorf("expected n2 to receive alert, got %d", len(n2.received))
	}
}

func TestFormatMessage(t *testing.T) {
	a := alert.Alert{
		JobName:    "cleanup",
		Level:      alert.LevelWarn,
		Message:    "overdue by 5m",
		OccurredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	msg := alert.FormatMessage(a)
	if !strings.Contains(msg, "cleanup") || !strings.Contains(msg, "WARN") {
		t.Errorf("unexpected format: %s", msg)
	}
}
