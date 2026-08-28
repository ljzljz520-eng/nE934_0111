package importer

import (
	"fmt"
	"strconv"
	"strings"
	"tastinginvite/internal/clock"
	"tastinginvite/internal/codec"
	"tastinginvite/internal/model"
	"tastinginvite/internal/repository"
	"tastinginvite/internal/validation"
	"time"
)

type Importer struct {
	Repo  *repository.Repository
	Clock clock.Clock
}

func (i Importer) Parse(line string) (model.ImportRow, error) {
	fields := strings.Split(line, "|")
	if len(fields) != 7 {
		return model.ImportRow{}, fmt.Errorf("expected 7 fields")
	}
	return model.ImportRow{ExternalID: fields[0], Title: fields[1], Host: fields[2], Venue: fields[3], Start: fields[4], End: fields[5], Capacity: fields[6]}, nil
}

func (i Importer) Convert(row model.ImportRow) (model.Record, error) {
	if issues := validation.ImportRow(row); len(issues) > 0 {
		return model.Record{}, issues[0]
	}
	start, err := time.Parse(time.RFC3339, row.Start)
	if err != nil {
		return model.Record{}, fmt.Errorf("start: %w", err)
	}
	end, err := time.Parse(time.RFC3339, row.End)
	if err != nil {
		return model.Record{}, fmt.Errorf("end: %w", err)
	}
	capacity, err := strconv.Atoi(row.Capacity)
	if err != nil {
		return model.Record{}, fmt.Errorf("capacity: %w", err)
	}
	return model.Record{ID: row.ExternalID, Title: row.Title, Host: row.Host, Venue: row.Venue, StartAt: start, EndAt: end, Capacity: capacity, Tags: codec.SplitTags(row.Tags)}, nil
}

func (i Importer) Run(lines []string, actor string) model.ImportReport {
	report := model.ImportReport{Errors: make([]string, 0), IDs: make([]string, 0)}
	for lineNumber, line := range lines {
		row, err := i.Parse(line)
		if err != nil {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d: %v", lineNumber+1, err))
			continue
		}
		record, err := i.Convert(row)
		if err != nil {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d: %v", lineNumber+1, err))
			continue
		}
		if _, err = i.Repo.Create(record, actor); err != nil {
			report.Rejected++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d: %v", lineNumber+1, err))
			continue
		}
		report.Imported++
		report.IDs = append(report.IDs, record.ID)
	}
	return report
}
