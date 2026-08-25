package schedule

import (
	"fmt"
	"sort"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"time"
)

type Window struct {
	Start time.Time
	End   time.Time
	Label string
}

type Planner struct {
	Clock   clock.Clock
	windows []Window
}

func New(c clock.Clock) *Planner { return &Planner{Clock: c, windows: make([]Window, 0)} }

func (p *Planner) AddWindow(window Window) error {
	if window.Start.IsZero() || window.End.IsZero() {
		return fmt.Errorf("window requires bounds")
	}
	if !window.End.After(window.Start) {
		return fmt.Errorf("window end must follow start")
	}
	if window.Start.Before(p.Clock.Now()) {
		return fmt.Errorf("window starts in the past")
	}
	for _, existing := range p.windows {
		if window.Start.Before(existing.End) && existing.Start.Before(window.End) {
			return fmt.Errorf("window overlaps %s", existing.Label)
		}
	}
	p.windows = append(p.windows, window)
	sort.Slice(p.windows, func(i, j int) bool { return p.windows[i].Start.Before(p.windows[j].Start) })
	return nil
}

func (p *Planner) Available(start, end time.Time) bool {
	if !end.After(start) {
		return false
	}
	for _, window := range p.windows {
		if start.Before(window.End) && window.Start.Before(end) {
			return false
		}
	}
	return true
}

func (p *Planner) Next(after time.Time) (Window, bool) {
	for _, window := range p.windows {
		if window.Start.After(after) {
			return window, true
		}
	}
	return Window{}, false
}

func (p *Planner) Windows() []Window {
	out := make([]Window, len(p.windows))
	copy(out, p.windows)
	return out
}

func Duration(record model.Record) time.Duration {
	if record.EndAt.Before(record.StartAt) {
		return 0
	}
	return record.EndAt.Sub(record.StartAt)
}

func RemainingCapacity(record model.Record, registered int) int {
	remaining := record.Capacity - registered
	if remaining < 0 {
		return 0
	}
	return remaining
}

func CanRegister(record model.Record, registered int, at time.Time) error {
	if record.Status != "published" {
		return fmt.Errorf("invitation is not published")
	}
	if registered >= record.Capacity {
		return fmt.Errorf("capacity reached")
	}
	if at.Before(record.StartAt) || at.After(record.EndAt) {
		return fmt.Errorf("registration is outside event window")
	}
	return nil
}

func NormalizeRange(start, end time.Time) (time.Time, time.Time, error) {
	start, end = clock.Normalize(start), clock.Normalize(end)
	if start.IsZero() || end.IsZero() {
		return start, end, fmt.Errorf("range requires times")
	}
	if !end.After(start) {
		return start, end, fmt.Errorf("range is empty")
	}
	return start, end, nil
}

func DayBuckets(records []model.Record) map[string][]model.Record {
	result := map[string][]model.Record{}
	for _, record := range records {
		key := record.StartAt.Format("2006-01-02")
		result[key] = append(result[key], record)
	}
	return result
}
