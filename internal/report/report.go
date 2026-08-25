package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"tastinginvite/internal/model"
)

func InvitationCSV(records []model.Record) (string, error) {
	sorted := append([]model.Record(nil), records...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	var b strings.Builder
	w := csv.NewWriter(&b)
	if err := w.Write([]string{"id", "title", "host", "venue", "status", "capacity", "version"}); err != nil {
		return "", err
	}
	for _, record := range sorted {
		if err := w.Write([]string{record.ID, record.Title, record.Host, record.Venue, record.Status, strconv.Itoa(record.Capacity), strconv.Itoa(record.Version)}); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return b.String(), nil
}

func InvitationText(record model.Record) string {
	return fmt.Sprintf("%s | %s | %s | %s | %s | capacity %d | v%d", record.ID, record.Title, record.Host, record.Venue, record.Status, record.Capacity, record.Version)
}

func ParseCSV(input io.Reader) ([]model.ImportRow, error) {
	reader := csv.NewReader(input)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	if len(header) < 3 {
		return nil, fmt.Errorf("header too short")
	}
	rows := make([]model.ImportRow, 0)
	for {
		values, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
		if len(values) < 7 {
			return nil, fmt.Errorf("row too short")
		}
		rows = append(rows, model.ImportRow{ExternalID: values[0], Title: values[1], Host: values[2], Venue: values[3], Start: values[4], End: values[5], Capacity: values[6]})
	}
	return rows, nil
}
