package flow026

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/model"
	"tastinginvite/internal/report"
	"tastinginvite/internal/repository"
)

type Handler struct {
	Repo     *repository.Repository
	Clock    clock.Clock
	mu       sync.Mutex
	previous *model.ExportResult
}

func New(repo *repository.Repository, c clock.Clock) *Handler { return &Handler{Repo: repo, Clock: c} }

func (h *Handler) Export(ctx context.Context, id string) (model.ExportResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	record, err := h.Repo.Get(id)
	if err != nil {
		return model.ExportResult{}, err
	}
	result := model.ExportResult{RecordID: id, Title: record.Title, Status: record.Status, Version: record.Version, GeneratedAt: clock.Normalize(h.Clock.Now()), Rows: []string{report.InvitationText(record)}}
	select {
	case <-ctx.Done():
		result.Cancelled = true
		if h.previous != nil {
			stale := *h.previous
			stale.Cancelled = true
			return stale, nil
		}
		return result, ctx.Err()
	default:
	}
	h.previous = &result
	return result, nil
}

func (h *Handler) ExportPage(ctx context.Context, ids []string) ([]model.ExportResult, error) {
	results := make([]model.ExportResult, 0, len(ids))
	for _, id := range ids {
		result, err := h.Export(ctx, id)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (h *Handler) Summary(ctx context.Context, ids []string) (string, error) {
	pages, err := h.ExportPage(ctx, ids)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(pages))
	for _, page := range pages {
		parts = append(parts, fmt.Sprintf("%s:%s", page.RecordID, page.Status))
	}
	return strings.Join(parts, ","), nil
}
