package audit

import (
	"fmt"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/store"
)

type Logger struct {
	Store *store.Store
	Clock clock.Clock
}

func (l Logger) Record(recordID, action, actor, detail string) (model.AuditEvent, error) {
	event := model.AuditEvent{ID: fmt.Sprintf("audit-%s-%d", recordID, l.Clock.Now().UnixNano()), RecordID: recordID, Action: action, Actor: actor, Detail: detail, At: clock.Normalize(l.Clock.Now())}
	return event, l.Store.PutAudit(event)
}

func (l Logger) History(recordID string) ([]model.AuditEvent, error) {
	return l.Store.ListAudits(recordID)
}

func Summarize(events []model.AuditEvent) map[string]int {
	result := map[string]int{}
	for _, event := range events {
		result[event.Action]++
	}
	return result
}
