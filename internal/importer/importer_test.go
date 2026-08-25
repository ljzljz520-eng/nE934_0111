package importer

import (
	"path/filepath"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/store"
	"testing"
	"time"
)

func TestImporterRunReport(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	now := time.Date(2032, 2, 2, 9, 0, 0, 0, time.UTC)
	i := Importer{Repo: repository.New(s, clock.Fixed{At: now}), Clock: clock.Fixed{At: now}}
	report := i.Run([]string{"ext-1|Cellar|Mina|Hall|2032-02-03T10:00:00Z|2032-02-03T12:00:00Z|10", "broken"}, "importer")
	if report.Imported != 1 || report.Rejected != 1 || len(report.IDs) != 1 {
		t.Fatalf("report=%+v", report)
	}
}

func TestImporterParseRejectsWrongShape(t *testing.T) {
	if _, err := (Importer{}).Parse("one|two"); err == nil {
		t.Fatal("expected parse error")
	}
}
