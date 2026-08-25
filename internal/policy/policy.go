package policy

import (
	"fmt"
	"strings"
	"tastinginvite/internal/model"
)

type Role string

const (
	RoleCoordinator Role = "coordinator"
	RoleReviewer    Role = "reviewer"
	RoleHost        Role = "host"
	RoleGuest       Role = "guest"
)

type Principal struct {
	ID          string
	Role        Role
	DisplayName string
	Active      bool
}

func Check(principal Principal, action string, record model.Record) error {
	if !principal.Active {
		return fmt.Errorf("principal is inactive")
	}
	action = strings.ToLower(strings.TrimSpace(action))
	switch principal.Role {
	case RoleCoordinator:
		return coordinatorAction(action, record)
	case RoleReviewer:
		return reviewerAction(action, record)
	case RoleHost:
		return hostAction(action, record, principal)
	case RoleGuest:
		if action == "view" && record.Status == "published" {
			return nil
		}
		return fmt.Errorf("guest cannot %s", action)
	default:
		return fmt.Errorf("unknown role %s", principal.Role)
	}
}

func coordinatorAction(action string, record model.Record) error {
	allowed := map[string]bool{"create": true, "update": true, "archive": true, "view": true, "import": true, "export": true}
	if !allowed[action] {
		return fmt.Errorf("coordinator cannot %s", action)
	}
	if action == "archive" && record.Status == "draft" {
		return fmt.Errorf("draft cannot be archived")
	}
	return nil
}

func reviewerAction(action string, record model.Record) error {
	if action == "review" && record.Status == "submitted" {
		return nil
	}
	if action == "view" && record.Status != "archived" {
		return nil
	}
	return fmt.Errorf("reviewer cannot %s for %s", action, record.Status)
}

func hostAction(action string, record model.Record, principal Principal) error {
	if action == "view" && record.Host == principal.DisplayName {
		return nil
	}
	if action == "update" && record.Status == "draft" && record.Host == principal.DisplayName {
		return nil
	}
	return fmt.Errorf("host cannot %s this invitation", action)
}

func Roles() []Role { return []Role{RoleCoordinator, RoleReviewer, RoleHost, RoleGuest} }

func IsStaff(role Role) bool { return role == RoleCoordinator || role == RoleReviewer }

func ValidatePrincipal(p Principal) error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("principal id required")
	}
	if strings.TrimSpace(p.DisplayName) == "" {
		return fmt.Errorf("display name required")
	}
	for _, role := range Roles() {
		if p.Role == role {
			return nil
		}
	}
	return fmt.Errorf("invalid role")
}

func Scope(principal Principal, records []model.Record) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range records {
		if principal.Role == RoleCoordinator || principal.Role == RoleReviewer || record.Host == principal.DisplayName {
			out = append(out, record)
		}
	}
	return out
}
