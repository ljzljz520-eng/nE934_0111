package repository

import (
	"fmt"
	"strings"
	"tastinginvite/internal/audit"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/store"
	"tastinginvite/internal/validation"
)

type Repository struct {
	Store *store.Store
	Clock clock.Clock
	Audit audit.Logger
}

func New(s *store.Store, c clock.Clock) *Repository {
	return &Repository{Store: s, Clock: c, Audit: audit.Logger{Store: s, Clock: c}}
}

func (r *Repository) Create(record model.Record, actor string) (model.Record, error) {
	record.CreatedAt = clock.Normalize(r.Clock.Now())
	record.UpdatedAt = record.CreatedAt
	record.Status = "draft"
	record.Version = 1
	if issues := validation.Record(record, r.Clock.Now()); len(issues) > 0 {
		return record, issues[0]
	}
	if err := r.Store.PutRecord(record); err != nil {
		return record, err
	}
	_, err := r.Audit.Record(record.ID, "created", actor, record.Title)
	return record, err
}

func (r *Repository) Get(id string) (model.Record, error) { return r.Store.GetRecord(id) }

func (r *Repository) Save(record model.Record, actor string) error {
	if err := validation.Status(record.Status); err != nil {
		return err
	}
	current, err := r.Get(record.ID)
	if err != nil {
		return err
	}
	if record.Version != current.Version {
		return fmt.Errorf("version conflict: have %d want %d", record.Version, current.Version)
	}
	record.Version++
	record.UpdatedAt = clock.Normalize(r.Clock.Now())
	if err := r.Store.PutRecord(record); err != nil {
		return err
	}
	_, err = r.Audit.Record(record.ID, "updated", actor, fmt.Sprintf("version=%d", record.Version))
	return err
}

func (r *Repository) Transition(id, status, actor string) (model.Record, error) {
	record, err := r.Get(id)
	if err != nil {
		return record, err
	}
	if err := validation.Transition(record.Status, status); err != nil {
		return record, err
	}
	record.Status = status
	if err := r.Save(record, actor); err != nil {
		return record, err
	}
	return record, nil
}

func (r *Repository) Search(filter model.InvitationFilter) ([]model.Record, error) {
	records, err := r.Store.ListRecords()
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	result := make([]model.Record, 0)
	for _, record := range records {
		if filter.Status != "" && record.Status != filter.Status {
			continue
		}
		if filter.Host != "" && !strings.EqualFold(record.Host, filter.Host) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(record.Title+" "+record.Description+" "+record.Venue), query) {
			continue
		}
		if filter.Tag != "" && !contains(record.Tags, filter.Tag) {
			continue
		}
		if filter.From != nil && record.StartAt.Before(*filter.From) {
			continue
		}
		if filter.To != nil && record.StartAt.After(*filter.To) {
			continue
		}
		result = append(result, record)
	}
	if filter.Offset > len(result) {
		return []model.Record{}, nil
	}
	result = result[filter.Offset:]
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func (r *Repository) Archive(id, actor string) error {
	_, err := r.Transition(id, "archived", actor)
	return err
}

func (r *Repository) AddAttachment(a model.Attachment, actor string) error {
	if a.ID == "" || a.RecordID == "" || len(a.Content) == 0 {
		return fmt.Errorf("attachment is incomplete")
	}
	if _, err := r.Get(a.RecordID); err != nil {
		return err
	}
	if err := r.Store.PutAttachment(a); err != nil {
		return err
	}
	_, err := r.Audit.Record(a.RecordID, "attachment_added", actor, a.Name)
	return err
}

func (r *Repository) Workflows(recordID string) ([]model.Workflow, error) {
	return r.Store.ListWorkflows(recordID)
}
