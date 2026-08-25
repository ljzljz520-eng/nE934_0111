package registration

import (
	"fmt"
	"sort"
	"strings"
	"tastinginvite/internal/model"
	"time"
)

type State string

const (
	StateReserved   State = "reserved"
	StateConfirmed  State = "confirmed"
	StateCancelled  State = "cancelled"
	StateWaitlisted State = "waitlisted"
)

type Guest struct {
	ID        string
	Name      string
	Email     string
	Phone     string
	Dietary   string
	Notes     string
	CreatedAt time.Time
}

type Entry struct {
	ID          string
	RecordID    string
	Guest       Guest
	State       State
	Seats       int
	ReservedAt  time.Time
	ConfirmedAt *time.Time
	CancelledAt *time.Time
	Source      string
}

type Manager struct {
	entries  map[string]Entry
	byRecord map[string][]string
}

func NewManager() *Manager {
	return &Manager{entries: map[string]Entry{}, byRecord: map[string][]string{}}
}

func (m *Manager) Reserve(record model.Record, guest Guest, seats int, at time.Time) (Entry, error) {
	if err := ValidateGuest(guest); err != nil {
		return Entry{}, err
	}
	if seats < 1 {
		return Entry{}, fmt.Errorf("seats must be positive")
	}
	if record.Status != "published" {
		return Entry{}, fmt.Errorf("record is not published")
	}
	if at.Before(record.StartAt) || at.After(record.EndAt) {
		return Entry{}, fmt.Errorf("reservation is outside event window")
	}
	if m.Seats(record.ID)+seats > record.Capacity {
		return Entry{}, fmt.Errorf("capacity reached")
	}
	id := fmt.Sprintf("reg-%s-%s", record.ID, guest.ID)
	if _, exists := m.entries[id]; exists {
		return Entry{}, fmt.Errorf("guest already registered")
	}
	entry := Entry{ID: id, RecordID: record.ID, Guest: guest, State: StateReserved, Seats: seats, ReservedAt: at, Source: "invite-page"}
	m.entries[id] = entry
	m.byRecord[record.ID] = append(m.byRecord[record.ID], id)
	return entry, nil
}

func (m *Manager) Confirm(id string, at time.Time) (Entry, error) {
	entry, ok := m.entries[id]
	if !ok {
		return Entry{}, fmt.Errorf("registration not found")
	}
	if entry.State != StateReserved {
		return entry, fmt.Errorf("registration is %s", entry.State)
	}
	entry.State = StateConfirmed
	entry.ConfirmedAt = &at
	m.entries[id] = entry
	return entry, nil
}

func (m *Manager) Cancel(id string, at time.Time) (Entry, error) {
	entry, ok := m.entries[id]
	if !ok {
		return Entry{}, fmt.Errorf("registration not found")
	}
	if entry.State == StateCancelled {
		return entry, nil
	}
	if entry.State == StateConfirmed && at.After(entry.ReservedAt.Add(24*time.Hour)) {
		return entry, fmt.Errorf("confirmed registration is locked")
	}
	entry.State = StateCancelled
	entry.CancelledAt = &at
	m.entries[id] = entry
	return entry, nil
}

func (m *Manager) Waitlist(record model.Record, guest Guest, seats int, at time.Time) (Entry, error) {
	if err := ValidateGuest(guest); err != nil {
		return Entry{}, err
	}
	if seats < 1 {
		return Entry{}, fmt.Errorf("seats must be positive")
	}
	id := fmt.Sprintf("wait-%s-%s", record.ID, guest.ID)
	entry := Entry{ID: id, RecordID: record.ID, Guest: guest, State: StateWaitlisted, Seats: seats, ReservedAt: at, Source: "invite-page"}
	m.entries[id] = entry
	m.byRecord[record.ID] = append(m.byRecord[record.ID], id)
	return entry, nil
}

func (m *Manager) Promote(record model.Record, at time.Time) (Entry, error) {
	ids := m.byRecord[record.ID]
	for _, id := range ids {
		entry := m.entries[id]
		if entry.State == StateWaitlisted && m.Seats(record.ID)+entry.Seats <= record.Capacity {
			entry.State = StateReserved
			entry.ReservedAt = at
			m.entries[id] = entry
			return entry, nil
		}
	}
	return Entry{}, fmt.Errorf("no waitlisted guest fits")
}

func (m *Manager) Get(id string) (Entry, bool) { entry, ok := m.entries[id]; return entry, ok }

func (m *Manager) List(recordID string) []Entry {
	ids := m.byRecord[recordID]
	out := make([]Entry, 0, len(ids))
	for _, id := range ids {
		if entry, ok := m.entries[id]; ok {
			out = append(out, entry)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (m *Manager) Seats(recordID string) int {
	total := 0
	for _, entry := range m.List(recordID) {
		if entry.State == StateReserved || entry.State == StateConfirmed {
			total += entry.Seats
		}
	}
	return total
}

func (m *Manager) Counts(recordID string) map[State]int {
	counts := map[State]int{}
	for _, entry := range m.List(recordID) {
		counts[entry.State]++
	}
	return counts
}

func (m *Manager) ReleaseExpired(now time.Time, ttl time.Duration) []Entry {
	released := make([]Entry, 0)
	for id, entry := range m.entries {
		if entry.State == StateReserved && now.Sub(entry.ReservedAt) > ttl {
			entry.State = StateCancelled
			entry.CancelledAt = &now
			m.entries[id] = entry
			released = append(released, entry)
		}
	}
	return released
}

func ValidateGuest(guest Guest) error {
	if strings.TrimSpace(guest.ID) == "" {
		return fmt.Errorf("guest id required")
	}
	if strings.TrimSpace(guest.Name) == "" {
		return fmt.Errorf("guest name required")
	}
	if !strings.Contains(guest.Email, "@") {
		return fmt.Errorf("guest email invalid")
	}
	return nil
}

func Confirmed(entries []Entry) []Entry {
	out := make([]Entry, 0)
	for _, entry := range entries {
		if entry.State == StateConfirmed {
			out = append(out, entry)
		}
	}
	return out
}

func Names(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Guest.Name)
	}
	sort.Strings(out)
	return out
}
