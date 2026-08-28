package clock

import "time"

type Clock interface{ Now() time.Time }

type Fixed struct{ At time.Time }

func (f Fixed) Now() time.Time { return f.At }

type System struct{}

func (System) Now() time.Time { return time.Now().UTC() }

func Normalize(t time.Time) time.Time {
	if t.IsZero() {
		return time.Unix(0, 0).UTC()
	}
	return t.UTC().Truncate(time.Second)
}

func Between(value, start, end time.Time) bool {
	value = Normalize(value)
	if !start.IsZero() && value.Before(Normalize(start)) {
		return false
	}
	if !end.IsZero() && value.After(Normalize(end)) {
		return false
	}
	return true
}

func SameDay(a, b time.Time) bool {
	a, b = Normalize(a), Normalize(b)
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}
