package query

import (
	"sort"
	"strings"
	"tastinginvite/internal/model"
)

type Index struct{ records []model.Record }

func New(records []model.Record) *Index { return &Index{records: records} }

func (i *Index) Titles(term string) []model.Record {
	term = strings.ToLower(strings.TrimSpace(term))
	out := make([]model.Record, 0)
	for _, record := range i.records {
		if term == "" || strings.Contains(strings.ToLower(record.Title), term) {
			out = append(out, record)
		}
	}
	return out
}

func (i *Index) Upcoming() []model.Record {
	out := append([]model.Record(nil), i.records...)
	sort.Slice(out, func(a, b int) bool { return out[a].StartAt.Before(out[b].StartAt) })
	return out
}

func (i *Index) ByStatus(status string) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range i.records {
		if record.Status == status {
			out = append(out, record)
		}
	}
	return out
}

func GroupByHost(records []model.Record) map[string][]model.Record {
	groups := map[string][]model.Record{}
	for _, record := range records {
		groups[record.Host] = append(groups[record.Host], record)
	}
	return groups
}
