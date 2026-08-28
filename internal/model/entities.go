package model

import "time"

type Record struct {
	ID          string
	Title       string
	Host        string
	Venue       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Capacity    int
	Status      string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ArchivedAt  *time.Time
	Tags        []string
}

type AuditEvent struct {
	ID       string
	RecordID string
	Action   string
	Actor    string
	Detail   string
	At       time.Time
}

type Workflow struct {
	ID          string
	RecordID    string
	Name        string
	Stage       string
	Owner       string
	DueAt       time.Time
	CompletedAt *time.Time
	Notes       string
}

type Attachment struct {
	ID        string
	RecordID  string
	Name      string
	MediaType string
	Content   []byte
	Checksum  string
	CreatedAt time.Time
}

type InvitationFilter struct {
	Query  string
	Status string
	Host   string
	Tag    string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

type ExportResult struct {
	RecordID    string
	Title       string
	Status      string
	Version     int
	GeneratedAt time.Time
	Rows        []string
	Cancelled   bool
}

type ImportRow struct {
	ExternalID string
	Title      string
	Host       string
	Venue      string
	Start      string
	End        string
	Capacity   string
	Tags       string
}

type ImportReport struct {
	Imported int
	Rejected int
	Errors   []string
	IDs      []string
}
