package flow026

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/store"
)

func newRepo(t *testing.T) *repository.Repository {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return repository.New(s, clock.Fixed{At: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)})
}

func makeRecord(id, title string) model.Record {
	base := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	return model.Record{
		ID: id, Title: title, Host: "host", Venue: "venue",
		Capacity: 10, StartAt: base, EndAt: base.Add(2 * time.Hour),
	}
}

// TestExportSecondPageReturnsCurrentRecord reproduces the bug where exporting the
// second invitation of a page, after the context has been cancelled, returned the
// stale first record's result instead of the current (second) record's result.
func TestExportSecondPageReturnsCurrentRecord(t *testing.T) {
	repo := newRepo(t)
	if _, err := repo.Create(makeRecord("rec-1", "Spring Tasting"), "tester"); err != nil {
		t.Fatalf("create rec-1: %v", err)
	}
	if _, err := repo.Create(makeRecord("rec-2", "Autumn Tasting"), "tester"); err != nil {
		t.Fatalf("create rec-2: %v", err)
	}
	h := New(repo, clock.Fixed{At: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)})

	// First export proceeds on a live context.
	first, err := h.Export(context.Background(), "rec-1")
	if err != nil {
		t.Fatalf("export rec-1: %v", err)
	}
	if first.RecordID != "rec-1" {
		t.Fatalf("first export id = %q, want rec-1", first.RecordID)
	}

	// Second export is made on a cancelled context. It must still report the
	// current record (rec-2), not the stale previous result (rec-1).
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	second, err := h.Export(ctx, "rec-2")
	if err != nil {
		t.Fatalf("export rec-2: %v", err)
	}
	if second.RecordID != "rec-2" {
		t.Fatalf("second export id = %q, want rec-2 (current record, not stale previous)", second.RecordID)
	}
	if !second.Cancelled {
		t.Fatalf("second export should be marked cancelled")
	}
	if second.Title != "Autumn Tasting" {
		t.Fatalf("second export title = %q, want Autumn Tasting", second.Title)
	}
}

// TestExportPageHonoursCurrentRecord verifies ExportPage against a context that
// is already cancelled: each result must correspond to its own record id.
func TestExportPageHonoursCurrentRecord(t *testing.T) {
	repo := newRepo(t)
	if _, err := repo.Create(makeRecord("p-1", "Page One"), "tester"); err != nil {
		t.Fatalf("create p-1: %v", err)
	}
	if _, err := repo.Create(makeRecord("p-2", "Page Two"), "tester"); err != nil {
		t.Fatalf("create p-2: %v", err)
	}
	h := New(repo, clock.Fixed{At: time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)})

	// Pre-seed h.previous by exporting p-1 on a live context.
	if _, err := h.Export(context.Background(), "p-1"); err != nil {
		t.Fatalf("seed export p-1: %v", err)
	}

	// Now export a fresh page on a cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := h.ExportPage(ctx, []string{"p-2"})
	if err != nil {
		t.Fatalf("export page: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].RecordID != "p-2" {
		t.Fatalf("page export id = %q, want p-2 (current record, not stale previous p-1)", results[0].RecordID)
	}
}
