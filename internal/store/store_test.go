package store

import (
	"path/filepath"
	"tastinginvite/internal/model"
	"testing"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invite.db")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := model.Record{ID: "reopen-1", Title: "Cellar Evening", Status: "draft", Version: 1}
	if err := s.PutRecord(record); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got, err := s.GetRecord("reopen-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != record.Title || got.Version != 1 {
		t.Fatalf("unexpected record: %+v", got)
	}
}

func TestStoreListsRelatedEntities(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "invite.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.PutWorkflow(model.Workflow{ID: "wf-1", RecordID: "r1", Name: "review"}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutAttachment(model.Attachment{ID: "a1", RecordID: "r1", Name: "menu", Content: []byte("x")}); err != nil {
		t.Fatal(err)
	}
	workflows, err := s.ListWorkflows("r1")
	if err != nil || len(workflows) != 1 {
		t.Fatalf("workflows=%v err=%v", workflows, err)
	}
	attachments, err := s.ListAttachments("r1")
	if err != nil || len(attachments) != 1 {
		t.Fatalf("attachments=%v err=%v", attachments, err)
	}
}
