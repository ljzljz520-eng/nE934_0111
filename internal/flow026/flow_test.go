package flow026

import (
	"context"
	"path/filepath"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/store"
	"testing"
	"time"
)

func exportFixture(t *testing.T) (*Handler, *repository.Repository, time.Time) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	now := time.Date(2033, 3, 3, 9, 0, 0, 0, time.UTC)
	c := clock.Fixed{At: now}
	r := repository.New(s, c)
	return New(r, c), r, now
}

func TestExportUsesCurrentRecord(t *testing.T) {
	h, r, now := exportFixture(t)
	for _, id := range []string{"e1", "e2"} {
		if _, err := r.Create(model.Record{ID: id, Title: id, Host: "H", Venue: "V", StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), Capacity: 5}, "host"); err != nil {
			t.Fatal(err)
		}
	}
	result, err := h.Export(context.Background(), "e2")
	if err != nil || result.RecordID != "e2" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func Test934BusinessRegression(t *testing.T) {
	h, r, now := exportFixture(t)
	first, err := r.Create(model.Record{ID: "first", Title: "First", Host: "H", Venue: "V", StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), Capacity: 5}, "host")
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.Create(model.Record{ID: "second", Title: "Second", Host: "H", Venue: "V", StartAt: now.Add(3 * time.Hour), EndAt: now.Add(4 * time.Hour), Capacity: 5}, "host")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.Export(context.Background(), first.ID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := h.Export(ctx, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.RecordID != second.ID || result.Title != second.Title {
		t.Fatalf("expected second invitation, got %+v", result)
	}
}
