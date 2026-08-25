package report

import (
	"strings"
	"tastinginvite/internal/model"
	"testing"
)

func TestInvitationCSVDeterministic(t *testing.T) {
	records := []model.Record{{ID: "b", Title: "B", Host: "H", Venue: "V", Status: "draft", Capacity: 2, Version: 1}, {ID: "a", Title: "A", Host: "H", Venue: "V", Status: "published", Capacity: 4, Version: 2}}
	value, err := InvitationCSV(records)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(value, "a,A,H,V,published,4,2") || strings.Index(value, "a,A") > strings.Index(value, "b,B") {
		t.Fatalf("csv=%s", value)
	}
}

func TestParseCSVRows(t *testing.T) {
	rows, err := ParseCSV(strings.NewReader("id,title,host,venue,start,end,capacity\nr1,Tasting,Mina,Hall,2030-01-01T10:00:00Z,2030-01-01T12:00:00Z,4\n"))
	if err != nil || len(rows) != 1 || rows[0].ExternalID != "r1" {
		t.Fatalf("rows=%v err=%v", rows, err)
	}
}
