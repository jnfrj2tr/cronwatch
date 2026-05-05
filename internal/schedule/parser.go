package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// NextRun returns the next scheduled run time after 'from' for the given cron expression.
func NextRun(expr string, from time.Time) (time.Time, error) {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	schedule, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}

	return schedule.Next(from), nil
}

// LastExpected returns the most recent scheduled run time before 'now' for the given cron expression.
// It steps backwards by checking previous intervals.
func LastExpected(expr string, now time.Time) (time.Time, error) {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	schedule, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}

	// Walk forward from a point in the past to find the last tick before now.
	// We search up to 2 years back to handle very infrequent schedules.
	searchFrom := now.Add(-2 * 365 * 24 * time.Hour)
	prev := searchFrom
	for {
		next := schedule.Next(prev)
		if next.IsZero() || next.After(now) {
			break
		}
		prev = next
	}

	if prev.Equal(searchFrom) {
		return time.Time{}, fmt.Errorf("no scheduled run found before %v for expression %q", now, expr)
	}

	return prev, nil
}

// IsValidExpression reports whether the given cron expression can be parsed.
func IsValidExpression(expr string) bool {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)
	_, err := parser.Parse(expr)
	return err == nil
}
