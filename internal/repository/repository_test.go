package repository

import (
	"path/filepath"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/store"
	"testing"
	"time"
)

func TestRepositoryCreateSearchAndVersion(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2030, 5, 1, 10, 0, 0, 0, time.UTC)
	r := New(s, clock.Fixed{At: now})
	created, err := r.Create(model.Record{ID: "r1", Title: "Rosé Salon", Host: "Ava", Venue: "Room 1", StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), Capacity: 12, Tags: []string{"rose"}}, "host")
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != "draft" || created.Version != 1 {
		t.Fatalf("created=%+v", created)
	}
	if _, err := r.Transition("r1", "submitted", "host"); err != nil {
		t.Fatal(err)
	}
	results, err := r.Search(model.InvitationFilter{Query: "salon", Tag: "ROSE"})
	if err != nil || len(results) != 1 {
		t.Fatalf("results=%v err=%v", results, err)
	}
}

func TestRepositoryRejectsStaleVersion(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2030, 5, 1, 10, 0, 0, 0, time.UTC)
	r := New(s, clock.Fixed{At: now})
	record, err := r.Create(model.Record{ID: "r1", Title: "Tasting", Host: "Ava", Venue: "Room", StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), Capacity: 2}, "host")
	if err != nil {
		t.Fatal(err)
	}
	record.Title = "Changed"
	if err := r.Save(record, "host"); err != nil {
		t.Fatal(err)
	}
	if err := r.Save(record, "host"); err == nil {
		t.Fatal("expected version conflict")
	}
}
