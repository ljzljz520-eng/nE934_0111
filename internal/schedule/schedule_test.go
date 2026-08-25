package schedule

import (
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"testing"
	"time"
)

func TestPlannerRejectsOverlapAndChecksCapacity(t *testing.T) {
	now := time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC)
	p := New(clock.Fixed{At: now})
	if err := p.AddWindow(Window{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Label: "one"}); err != nil {
		t.Fatal(err)
	}
	if err := p.AddWindow(Window{Start: now.Add(90 * time.Minute), End: now.Add(3 * time.Hour), Label: "two"}); err == nil {
		t.Fatal("expected overlap")
	}
	record := model.Record{Status: "published", Capacity: 4, StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour)}
	if err := CanRegister(record, 4, now.Add(90*time.Minute)); err == nil {
		t.Fatal("expected capacity error")
	}
}

func TestDayBuckets(t *testing.T) {
	records := []model.Record{{ID: "a", StartAt: time.Date(2030, 1, 1, 1, 0, 0, 0, time.UTC)}, {ID: "b", StartAt: time.Date(2030, 1, 1, 4, 0, 0, 0, time.UTC)}}
	if len(DayBuckets(records)["2030-01-01"]) != 2 {
		t.Fatal("expected same day bucket")
	}
}
