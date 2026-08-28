package insights

import (
	"fmt"
	"sort"
	"strings"
	"tastinginvite/internal/model"
	"time"
)

type Segment struct {
	Name     string
	Count    int
	Capacity int
	Share    float64
}

type TrendPoint struct {
	Day       time.Time
	Created   int
	Published int
	Archived  int
}

type Dashboard struct {
	Total     int
	Active    int
	Published int
	Pending   int
	Capacity  int
	Hosts     []string
	Segments  []Segment
}

func DashboardFor(records []model.Record) Dashboard {
	d := Dashboard{Hosts: make([]string, 0), Segments: make([]Segment, 0)}
	hosts := map[string]bool{}
	groups := map[string][]model.Record{}
	for _, record := range records {
		d.Total++
		d.Capacity += record.Capacity
		if record.Status != "archived" {
			d.Active++
		}
		if record.Status == "published" {
			d.Published++
		}
		if record.Status == "submitted" || record.Status == "approved" {
			d.Pending++
		}
		hosts[record.Host] = true
		groups[record.Status] = append(groups[record.Status], record)
	}
	for host := range hosts {
		d.Hosts = append(d.Hosts, host)
	}
	sort.Strings(d.Hosts)
	d.Segments = Segments(groups, d.Total)
	return d
}

func Segments(groups map[string][]model.Record, total int) []Segment {
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]Segment, 0, len(names))
	for _, name := range names {
		values := groups[name]
		capacity := 0
		for _, value := range values {
			capacity += value.Capacity
		}
		share := 0.0
		if total > 0 {
			share = float64(len(values)) / float64(total)
		}
		out = append(out, Segment{Name: name, Count: len(values), Capacity: capacity, Share: share})
	}
	return out
}

func Trend(records []model.Record, from, to time.Time) []TrendPoint {
	points := map[string]TrendPoint{}
	for _, record := range records {
		if record.CreatedAt.Before(from) || record.CreatedAt.After(to) {
			continue
		}
		day := record.CreatedAt.UTC().Truncate(24 * time.Hour)
		key := day.Format("2006-01-02")
		point := points[key]
		point.Day = day
		point.Created++
		if record.Status == "published" {
			point.Published++
		}
		if record.Status == "archived" {
			point.Archived++
		}
		points[key] = point
	}
	out := make([]TrendPoint, 0, len(points))
	for _, point := range points {
		out = append(out, point)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day.Before(out[j].Day) })
	return out
}

func Find(records []model.Record, phrase string) []model.Record {
	phrase = strings.ToLower(strings.TrimSpace(phrase))
	out := make([]model.Record, 0)
	for _, record := range records {
		haystack := strings.ToLower(strings.Join([]string{record.ID, record.Title, record.Host, record.Venue, record.Description}, " "))
		if phrase == "" || strings.Contains(haystack, phrase) {
			out = append(out, record)
		}
	}
	return out
}

func Summaries(records []model.Record) []string {
	out := make([]string, 0, len(records))
	for _, record := range records {
		label := record.Status
		if label == "published" {
			label = "open"
		}
		out = append(out, fmt.Sprintf("%s: %s (%s)", record.ID, record.Title, label))
	}
	sort.Strings(out)
	return out
}

func HostNames(records []model.Record) []string {
	seen := map[string]bool{}
	for _, record := range records {
		if strings.TrimSpace(record.Host) != "" {
			seen[record.Host] = true
		}
	}
	out := make([]string, 0, len(seen))
	for host := range seen {
		out = append(out, host)
	}
	sort.Strings(out)
	return out
}

func StatusCounts(records []model.Record) map[string]int {
	out := map[string]int{}
	for _, record := range records {
		out[record.Status]++
	}
	return out
}

func CapacityBands(records []model.Record) map[string]int {
	out := map[string]int{"small": 0, "medium": 0, "large": 0}
	for _, record := range records {
		switch {
		case record.Capacity < 20:
			out["small"]++
		case record.Capacity <= 80:
			out["medium"]++
		default:
			out["large"]++
		}
	}
	return out
}

func Upcoming(records []model.Record, now time.Time, limit int) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range records {
		if record.StartAt.Before(now) || record.Status == "archived" {
			continue
		}
		out = append(out, record)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartAt.Before(out[j].StartAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func Overdue(records []model.Record, now time.Time) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range records {
		if record.Status == "submitted" && record.UpdatedAt.Before(now.Add(-72*time.Hour)) {
			out = append(out, record)
		}
	}
	return out
}

func Completion(records []model.Record) float64 {
	if len(records) == 0 {
		return 0
	}
	done := 0
	for _, record := range records {
		if record.Status == "published" || record.Status == "archived" {
			done++
		}
	}
	return float64(done) / float64(len(records))
}

func AverageLead(records []model.Record) time.Duration {
	if len(records) == 0 {
		return 0
	}
	var total time.Duration
	for _, record := range records {
		if record.StartAt.After(record.CreatedAt) {
			total += record.StartAt.Sub(record.CreatedAt)
		}
	}
	return total / time.Duration(len(records))
}

func TagCloud(records []model.Record) []Segment {
	counts := map[string]int{}
	for _, record := range records {
		for _, tag := range record.Tags {
			counts[tag]++
		}
	}
	total := 0
	for _, count := range counts {
		total += count
	}
	groups := map[string][]model.Record{}
	for tag, count := range counts {
		groups[tag] = make([]model.Record, count)
	}
	return Segments(groups, total)
}

func MatchFilter(record model.Record, filter model.InvitationFilter) bool {
	if filter.Status != "" && record.Status != filter.Status {
		return false
	}
	if filter.Host != "" && !strings.EqualFold(record.Host, filter.Host) {
		return false
	}
	if filter.Tag != "" {
		found := false
		for _, tag := range record.Tags {
			if strings.EqualFold(tag, filter.Tag) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.From != nil && record.StartAt.Before(*filter.From) {
		return false
	}
	if filter.To != nil && record.StartAt.After(*filter.To) {
		return false
	}
	return true
}

func Filter(records []model.Record, filter model.InvitationFilter) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range records {
		if MatchFilter(record, filter) {
			out = append(out, record)
		}
	}
	return out
}

func Compare(a, b model.Record) int {
	if a.UpdatedAt.Before(b.UpdatedAt) {
		return -1
	}
	if a.UpdatedAt.After(b.UpdatedAt) {
		return 1
	}
	return strings.Compare(a.ID, b.ID)
}

func SortRecent(records []model.Record) []model.Record {
	out := append([]model.Record(nil), records...)
	sort.Slice(out, func(i, j int) bool { return Compare(out[i], out[j]) > 0 })
	return out
}

func DistinctTags(records []model.Record) []string {
	seen := map[string]bool{}
	for _, record := range records {
		for _, tag := range record.Tags {
			seen[tag] = true
		}
	}
	out := make([]string, 0, len(seen))
	for tag := range seen {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

func RecordLabel(record model.Record) string {
	if record.Title == "" {
		return record.ID
	}
	if record.Venue == "" {
		return record.Title
	}
	return record.Title + " at " + record.Venue
}

func IsVisible(record model.Record) bool { return record.Status == "published" }

func IsMutable(record model.Record) bool {
	return record.Status == "draft" || record.Status == "rejected"
}

func NeedsReview(record model.Record) bool { return record.Status == "submitted" }

func CanExport(record model.Record) bool {
	return record.Status != "draft" && record.Status != "rejected"
}

func Window(record model.Record) string {
	return record.StartAt.Format("2006-01-02 15:04") + " - " + record.EndAt.Format("15:04")
}
