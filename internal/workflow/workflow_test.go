package workflow

import (
	"path/filepath"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/store"
	"testing"
	"time"
)

func workflowFixture(t *testing.T) (*Engine, *repository.Repository, time.Time) {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	now := time.Date(2031, 1, 1, 9, 0, 0, 0, time.UTC)
	r := repository.New(s, clock.Fixed{At: now})
	return New(r, clock.Fixed{At: now}), r, now
}

func newWorkflowRecord(t *testing.T, r *repository.Repository, now time.Time, id string) {
	t.Helper()
	_, err := r.Create(model.Record{ID: id, Title: "Invitation " + id, Host: "Host", Venue: "Gallery", StartAt: now.Add(time.Hour), EndAt: now.Add(2 * time.Hour), Capacity: 8}, "coordinator")
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowCreateReviewArchive(t *testing.T) {
	e, r, now := workflowFixture(t)
	newWorkflowRecord(t, r, now, "w1")
	if _, err := e.Submit("w1", "host"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Review("w1", "reviewer", true, "clear brief"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Publish("w1", "coordinator"); err != nil {
		t.Fatal(err)
	}
	got, err := e.Archive("w1", "coordinator")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "archived" {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	e, r, now := workflowFixture(t)
	newWorkflowRecord(t, r, now, "w2")
	if _, err := e.Submit("w2", "host"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Review("w2", "reviewer", true, "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Publish("w2", "coordinator"); err != nil {
		t.Fatal(err)
	}
	results, err := r.Search(model.InvitationFilter{Status: "published"})
	if err != nil || len(results) != 1 {
		t.Fatalf("results=%v err=%v", results, err)
	}
}

func TestWorkflowImportReport(t *testing.T) {
	e, r, now := workflowFixture(t)
	newWorkflowRecord(t, r, now, "w3")
	if _, err := e.Assign("w3", "menu-check", "reviewer", now.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	completed, err := e.Complete("w3", "menu-check", now.Add(2*time.Hour), "checked")
	if err != nil {
		t.Fatal(err)
	}
	if completed.Stage != "completed" || completed.Notes != "checked" {
		t.Fatalf("completed=%+v", completed)
	}
}
