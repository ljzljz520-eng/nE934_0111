package codec

import (
	"encoding/json"
	"fmt"
	"strings"
	"tastinginvite/internal/model"
)

func EncodeRecord(r model.Record) ([]byte, error) { return json.Marshal(r) }

func DecodeRecord(data []byte) (model.Record, error) {
	var r model.Record
	if err := json.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("decode record: %w", err)
	}
	return r, nil
}

func EncodeAudit(e model.AuditEvent) ([]byte, error) { return json.Marshal(e) }

func DecodeAudit(data []byte) (model.AuditEvent, error) {
	var e model.AuditEvent
	if err := json.Unmarshal(data, &e); err != nil {
		return e, fmt.Errorf("decode audit: %w", err)
	}
	return e, nil
}

func EncodeWorkflow(w model.Workflow) ([]byte, error) { return json.Marshal(w) }

func DecodeWorkflow(data []byte) (model.Workflow, error) {
	var w model.Workflow
	if err := json.Unmarshal(data, &w); err != nil {
		return w, fmt.Errorf("decode workflow: %w", err)
	}
	return w, nil
}

func EncodeAttachment(a model.Attachment) ([]byte, error) { return json.Marshal(a) }

func DecodeAttachment(data []byte) (model.Attachment, error) {
	var a model.Attachment
	if err := json.Unmarshal(data, &a); err != nil {
		return a, fmt.Errorf("decode attachment: %w", err)
	}
	return a, nil
}

func SplitTags(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || seen[part] {
			continue
		}
		seen[part] = true
		result = append(result, part)
	}
	return result
}

func JoinTags(tags []string) string { return strings.Join(tags, ",") }
