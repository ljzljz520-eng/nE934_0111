package flow026

import (
	"context"
	"fmt"
	"tastinginvite/internal/model"
)

func ValidateExport(result model.ExportResult) error {
	if result.RecordID == "" {
		return fmt.Errorf("record id missing")
	}
	if result.Title == "" {
		return fmt.Errorf("title missing")
	}
	if result.Cancelled && len(result.Rows) == 0 {
		return fmt.Errorf("cancelled result has no context")
	}
	return nil
}

func Cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func RenderRows(result model.ExportResult) []string {
	rows := make([]string, len(result.Rows))
	copy(rows, result.Rows)
	return rows
}

func StatusLabel(status string) string {
	switch status {
	case "approved":
		return "Ready for guests"
	case "published":
		return "Visible to invitees"
	case "archived":
		return "Archived"
	case "rejected":
		return "Needs revision"
	default:
		return "In preparation"
	}
}
