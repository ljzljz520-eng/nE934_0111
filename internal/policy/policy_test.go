package policy

import (
	"tastinginvite/internal/model"
	"testing"
)

func TestPolicyScopesHosts(t *testing.T) {
	principal := Principal{ID: "p1", Role: RoleHost, DisplayName: "Mina", Active: true}
	records := []model.Record{{ID: "one", Host: "Mina"}, {ID: "two", Host: "Other"}}
	scoped := Scope(principal, records)
	if len(scoped) != 1 || scoped[0].ID != "one" {
		t.Fatalf("scoped=%v", scoped)
	}
	if err := Check(principal, "view", records[0]); err != nil {
		t.Fatal(err)
	}
	if err := Check(principal, "archive", records[0]); err == nil {
		t.Fatal("expected denial")
	}
}

func TestValidatePrincipal(t *testing.T) {
	if err := ValidatePrincipal(Principal{ID: "", Role: RoleGuest, DisplayName: "Guest", Active: true}); err == nil {
		t.Fatal("expected missing id")
	}
}
