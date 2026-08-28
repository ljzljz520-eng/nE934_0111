package insights

import (
	"sort"
	"strings"
	"tastinginvite/internal/model"
	"time"
)

type SortMode string

const (
	SortByDate     SortMode = "date"
	SortByTitle    SortMode = "title"
	SortByStatus   SortMode = "status"
	SortByCapacity SortMode = "capacity"
)

func SortRecords(records []model.Record, mode SortMode, descending bool) []model.Record {
	out := append([]model.Record(nil), records...)
	less := func(i, j int) bool {
		switch mode {
		case SortByTitle:
			return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
		case SortByStatus:
			if out[i].Status == out[j].Status {
				return out[i].ID < out[j].ID
			}
			return out[i].Status < out[j].Status
		case SortByCapacity:
			if out[i].Capacity == out[j].Capacity {
				return out[i].ID < out[j].ID
			}
			return out[i].Capacity < out[j].Capacity
		default:
			if out[i].StartAt.Equal(out[j].StartAt) {
				return out[i].ID < out[j].ID
			}
			return out[i].StartAt.Before(out[j].StartAt)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if descending {
			return less(j, i)
		}
		return less(i, j)
	})
	return out
}

func Page(records []model.Record, limit, offset int) []model.Record {
	if offset < 0 {
		offset = 0
	}
	if limit < 0 {
		limit = 0
	}
	if offset >= len(records) {
		return []model.Record{}
	}
	end := len(records)
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return append([]model.Record(nil), records[offset:end]...)
}

func Between(records []model.Record, start, end time.Time) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range records {
		if !start.IsZero() && record.StartAt.Before(start) {
			continue
		}
		if !end.IsZero() && record.StartAt.After(end) {
			continue
		}
		out = append(out, record)
	}
	return out
}

func StatusOptions(records []model.Record) []string {
	seen := map[string]bool{}
	for _, record := range records {
		seen[record.Status] = true
	}
	out := make([]string, 0, len(seen))
	for status := range seen {
		out = append(out, status)
	}
	sort.Strings(out)
	return out
}

func HostsWithCapacity(records []model.Record, minimum int) []string {
	totals := map[string]int{}
	for _, record := range records {
		totals[record.Host] += record.Capacity
	}
	out := make([]string, 0)
	for host, total := range totals {
		if total >= minimum {
			out = append(out, host)
		}
	}
	sort.Strings(out)
	return out
}

func CloneRecord(record model.Record) model.Record {
	copy := record
	copy.Tags = append([]string(nil), record.Tags...)
	if record.ArchivedAt != nil {
		value := *record.ArchivedAt
		copy.ArchivedAt = &value
	}
	return copy
}

func CloneRecords(records []model.Record) []model.Record {
	out := make([]model.Record, 0, len(records))
	for _, record := range records {
		out = append(out, CloneRecord(record))
	}
	return out
}
