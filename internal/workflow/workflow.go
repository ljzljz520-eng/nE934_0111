package workflow

import (
	"fmt"
	"tastinginvite/internal/audit"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/repository"
	"time"
)

type Engine struct {
	Repo  *repository.Repository
	Clock clock.Clock
	Audit audit.Logger
}

func New(repo *repository.Repository, c clock.Clock) *Engine {
	return &Engine{Repo: repo, Clock: c, Audit: audit.Logger{Store: repo.Store, Clock: c}}
}

func (e *Engine) Submit(id, actor string) (model.Record, error) {
	return e.Repo.Transition(id, "submitted", actor)
}

func (e *Engine) Review(id, actor string, approve bool, note string) (model.Record, error) {
	record, err := e.Repo.Get(id)
	if err != nil {
		return record, err
	}
	target := "rejected"
	if approve {
		target = "approved"
	}
	record, err = e.Repo.Transition(id, target, actor)
	if err != nil {
		return record, err
	}
	_, auditErr := e.Audit.Record(id, "reviewed", actor, fmt.Sprintf("approved=%t note=%s", approve, note))
	return record, auditErr
}

func (e *Engine) Publish(id, actor string) (model.Record, error) {
	return e.Repo.Transition(id, "published", actor)
}

func (e *Engine) Archive(id, actor string) (model.Record, error) {
	if err := e.Repo.Archive(id, actor); err != nil {
		return model.Record{}, err
	}
	return e.Repo.Get(id)
}

func (e *Engine) Assign(recordID, name, owner string, dueAt time.Time) (model.Workflow, error) {
	workflow := model.Workflow{ID: fmt.Sprintf("wf-%s-%s", recordID, name), RecordID: recordID, Name: name, Stage: "assigned", Owner: owner, DueAt: clock.Normalize(dueAt)}
	if _, err := e.Repo.Get(recordID); err != nil {
		return workflow, err
	}
	if err := e.Repo.Store.PutWorkflow(workflow); err != nil {
		return workflow, err
	}
	return workflow, nil
}

func (e *Engine) Complete(recordID, name string, completedAt time.Time, notes string) (model.Workflow, error) {
	workflows, err := e.Repo.Workflows(recordID)
	if err != nil {
		return model.Workflow{}, err
	}
	for _, workflow := range workflows {
		if workflow.Name == name {
			workflow.Stage = "completed"
			value := clock.Normalize(completedAt)
			workflow.CompletedAt = &value
			workflow.Notes = notes
			if err := e.Repo.Store.PutWorkflow(workflow); err != nil {
				return workflow, err
			}
			return workflow, nil
		}
	}
	return model.Workflow{}, fmt.Errorf("workflow %s not found", name)
}
