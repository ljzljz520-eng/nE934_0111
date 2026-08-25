package metrics

import (
	"sort"
	"tastinginvite/internal/model"
	"time"
)

type Snapshot struct {
	Total     int
	ByStatus  map[string]int
	ByHost    map[string]int
	Capacity  int
	Published int
	Archived  int
}

func Build(records []model.Record) Snapshot {
	snapshot := Snapshot{ByStatus: map[string]int{}, ByHost: map[string]int{}}
	for _, record := range records {
		snapshot.Total++
		snapshot.ByStatus[record.Status]++
		snapshot.ByHost[record.Host]++
		snapshot.Capacity += record.Capacity
		if record.Status == "published" {
			snapshot.Published++
		}
		if record.Status == "archived" {
			snapshot.Archived++
		}
	}
	return snapshot
}

func ConversionRate(snapshot Snapshot) float64 {
	if snapshot.Total == 0 {
		return 0
	}
	return float64(snapshot.Published) / float64(snapshot.Total)
}

func CapacityByHost(records []model.Record) map[string]int {
	result := map[string]int{}
	for _, record := range records {
		result[record.Host] += record.Capacity
	}
	return result
}

func StatusOrder(snapshot Snapshot) []string {
	keys := make([]string, 0, len(snapshot.ByStatus))
	for key := range snapshot.ByStatus {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func Recent(records []model.Record, since time.Time) []model.Record {
	out := make([]model.Record, 0)
	for _, record := range records {
		if !record.UpdatedAt.Before(since) {
			out = append(out, record)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

func AverageCapacity(records []model.Record) int {
	if len(records) == 0 {
		return 0
	}
	total := 0
	for _, record := range records {
		total += record.Capacity
	}
	return total / len(records)
}

func Tags(records []model.Record) map[string]int {
	result := map[string]int{}
	for _, record := range records {
		for _, tag := range record.Tags {
			result[tag]++
		}
	}
	return result
}
