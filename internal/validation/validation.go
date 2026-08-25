package validation

import (
	"fmt"
	"strings"
	"tastinginvite/internal/model"
	"time"
)

type Issue struct {
	Field   string
	Message string
}

func (i Issue) Error() string { return i.Field + ": " + i.Message }

func Record(r model.Record, now time.Time) []Issue {
	issues := make([]Issue, 0, 8)
	if strings.TrimSpace(r.ID) == "" {
		issues = append(issues, Issue{"id", "is required"})
	}
	if strings.TrimSpace(r.Title) == "" {
		issues = append(issues, Issue{"title", "is required"})
	}
	if len(r.Title) > 160 {
		issues = append(issues, Issue{"title", "must be 160 characters or fewer"})
	}
	if strings.TrimSpace(r.Host) == "" {
		issues = append(issues, Issue{"host", "is required"})
	}
	if strings.TrimSpace(r.Venue) == "" {
		issues = append(issues, Issue{"venue", "is required"})
	}
	if r.Capacity < 1 {
		issues = append(issues, Issue{"capacity", "must be positive"})
	}
	if r.Capacity > 10000 {
		issues = append(issues, Issue{"capacity", "exceeds room limit"})
	}
	if r.StartAt.IsZero() {
		issues = append(issues, Issue{"start_at", "is required"})
	}
	if r.EndAt.IsZero() {
		issues = append(issues, Issue{"end_at", "is required"})
	}
	if !r.StartAt.IsZero() && !r.EndAt.IsZero() && !r.EndAt.After(r.StartAt) {
		issues = append(issues, Issue{"end_at", "must follow start_at"})
	}
	if !r.StartAt.IsZero() && r.StartAt.Before(now.Add(-time.Minute)) {
		issues = append(issues, Issue{"start_at", "cannot be in the past"})
	}
	if len(r.Tags) > 12 {
		issues = append(issues, Issue{"tags", "too many tags"})
	}
	return issues
}

func Status(value string) error {
	for _, allowed := range []string{"draft", "submitted", "approved", "published", "archived", "rejected"} {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("invalid status %q", value)
}

func Transition(from, to string) error {
	if err := Status(from); err != nil {
		return err
	}
	if err := Status(to); err != nil {
		return err
	}
	valid := map[string]map[string]bool{
		"draft":     {"submitted": true, "rejected": true},
		"submitted": {"approved": true, "rejected": true, "draft": true},
		"approved":  {"published": true, "rejected": true},
		"published": {"archived": true},
		"rejected":  {"draft": true, "archived": true},
		"archived":  {},
	}
	if !valid[from][to] {
		return fmt.Errorf("cannot transition %s to %s", from, to)
	}
	return nil
}

func ImportRow(row model.ImportRow) []Issue {
	issues := make([]Issue, 0, 5)
	if strings.TrimSpace(row.ExternalID) == "" {
		issues = append(issues, Issue{"external_id", "is required"})
	}
	if strings.TrimSpace(row.Title) == "" {
		issues = append(issues, Issue{"title", "is required"})
	}
	if strings.TrimSpace(row.Host) == "" {
		issues = append(issues, Issue{"host", "is required"})
	}
	if strings.TrimSpace(row.Start) == "" {
		issues = append(issues, Issue{"start", "is required"})
	}
	if strings.TrimSpace(row.End) == "" {
		issues = append(issues, Issue{"end", "is required"})
	}
	return issues
}
