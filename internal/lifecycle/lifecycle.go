package lifecycle

import (
	"fmt"
	"sort"
	"tastinginvite/internal/model"
	"time"
)

type Event struct {
	ID       string
	RecordID string
	Name     string
	At       time.Time
	Actor    string
	Detail   string
}

type Timeline struct{ events []Event }

func New() *Timeline { return &Timeline{events: make([]Event, 0)} }

func (t *Timeline) Add(event Event) error {
	if event.ID == "" || event.RecordID == "" {
		return fmt.Errorf("event identity required")
	}
	if event.Name == "" {
		return fmt.Errorf("event name required")
	}
	t.events = append(t.events, event)
	sort.SliceStable(t.events, func(i, j int) bool { return t.events[i].At.Before(t.events[j].At) })
	return nil
}

func (t *Timeline) ForRecord(recordID string) []Event {
	out := make([]Event, 0)
	for _, event := range t.events {
		if event.RecordID == recordID {
			out = append(out, event)
		}
	}
	return out
}

func (t *Timeline) Latest(recordID string) (Event, bool) {
	events := t.ForRecord(recordID)
	if len(events) == 0 {
		return Event{}, false
	}
	return events[len(events)-1], true
}

func (t *Timeline) Since(recordID string, at time.Time) []Event {
	out := make([]Event, 0)
	for _, event := range t.events {
		if event.RecordID == recordID && !event.At.Before(at) {
			out = append(out, event)
		}
	}
	return out
}

func (t *Timeline) Names(recordID string) []string {
	events := t.ForRecord(recordID)
	names := make([]string, 0, len(events))
	for _, event := range events {
		names = append(names, event.Name)
	}
	return names
}

func Build(record model.Record, audits []model.AuditEvent) *Timeline {
	timeline := New()
	for _, audit := range audits {
		_ = timeline.Add(Event{ID: audit.ID, RecordID: record.ID, Name: audit.Action, At: audit.At, Actor: audit.Actor, Detail: audit.Detail})
	}
	return timeline
}

func CanArchive(record model.Record, at time.Time) error {
	if record.Status != "published" && record.Status != "rejected" {
		return fmt.Errorf("status %s cannot be archived", record.Status)
	}
	if at.Before(record.EndAt) {
		return fmt.Errorf("event has not ended")
	}
	return nil
}

func Age(record model.Record, at time.Time) time.Duration { return at.Sub(record.CreatedAt) }

func Sort(events []Event) []Event {
	out := append([]Event(nil), events...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}
