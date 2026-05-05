package schedule_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/schedule"
)

func mustTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestNextRun_Valid(t *testing.T) {
	// Every hour at :00
	next, err := schedule.NextRun("0 * * * *", mustTime("2024-01-15 10:30"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := mustTime("2024-01-15 11:00")
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestNextRun_InvalidExpr(t *testing.T) {
	_, err := schedule.NextRun("not-a-cron", time.Now())
	if err == nil {
		t.Error("expected error for invalid cron expression")
	}
}

func TestLastExpected_Valid(t *testing.T) {
	// Every day at midnight
	last, err := schedule.LastExpected("0 0 * * *", mustTime("2024-01-15 10:30"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := mustTime("2024-01-15 00:00")
	if !last.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, last)
	}
}

func TestLastExpected_InvalidExpr(t *testing.T) {
	_, err := schedule.LastExpected("bad expr", time.Now())
	if err == nil {
		t.Error("expected error for invalid cron expression")
	}
}

func TestIsValidExpression(t *testing.T) {
	tests := []struct {
		expr  string
		valid bool
	}{
		{"* * * * *", true},
		{"0 9 * * 1-5", true},
		{"30 6 1 1 *", true},
		{"not-valid", false},
		{"", false},
		{"0 0 0 0 0", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got := schedule.IsValidExpression(tt.expr)
			if got != tt.valid {
				t.Errorf("IsValidExpression(%q) = %v, want %v", tt.expr, got, tt.valid)
			}
		})
	}
}
