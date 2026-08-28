package analytics

import (
	"math"
	"sort"
	"tastinginvite/internal/model"
	"time"
)

type VenueScore struct {
	Venue    string
	Events   int
	Capacity int
	Average  int
}

type Forecast struct {
	Day      string
	Events   int
	Capacity int
	Load     float64
}

func RankVenues(records []model.Record) []VenueScore {
	grouped := map[string][]model.Record{}
	for _, record := range records {
		grouped[record.Venue] = append(grouped[record.Venue], record)
	}
	out := make([]VenueScore, 0, len(grouped))
	for venue, values := range grouped {
		total := 0
		for _, value := range values {
			total += value.Capacity
		}
		out = append(out, VenueScore{Venue: venue, Events: len(values), Capacity: total, Average: total / len(values)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Capacity == out[j].Capacity {
			return out[i].Venue < out[j].Venue
		}
		return out[i].Capacity > out[j].Capacity
	})
	return out
}

func ForecastDays(records []model.Record) []Forecast {
	grouped := map[string][]model.Record{}
	for _, record := range records {
		key := record.StartAt.Format("2006-01-02")
		grouped[key] = append(grouped[key], record)
	}
	out := make([]Forecast, 0, len(grouped))
	for day, values := range grouped {
		capacity := 0
		for _, value := range values {
			capacity += value.Capacity
		}
		load := float64(capacity) / float64(len(values))
		out = append(out, Forecast{Day: day, Events: len(values), Capacity: capacity, Load: math.Round(load*100) / 100})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day < out[j].Day })
	return out
}

func Utilization(record model.Record, confirmed int) float64 {
	if record.Capacity <= 0 {
		return 0
	}
	value := float64(confirmed) / float64(record.Capacity)
	if value > 1 {
		return 1
	}
	return value
}

func Peak(records []model.Record) (time.Time, int) {
	counts := map[string]int{}
	for _, record := range records {
		counts[record.StartAt.Format("2006-01-02")]++
	}
	day := ""
	count := 0
	for value, amount := range counts {
		if amount > count || (amount == count && value < day) {
			day, count = value, amount
		}
	}
	if day == "" {
		return time.Time{}, 0
	}
	parsed, _ := time.Parse("2006-01-02", day)
	return parsed, count
}

func StatusMix(records []model.Record) map[string]float64 {
	total := len(records)
	result := map[string]float64{}
	if total == 0 {
		return result
	}
	for _, record := range records {
		result[record.Status] += 1 / float64(total)
	}
	return result
}
