package calendar

import "time"

type Day struct {
	Date    time.Time
	InMonth bool
}

// Weeks returns the month as full Sunday-to-Saturday rows, padded with days
// from the neighbouring months.
func Weeks(year int, month time.Month) [][]Day {
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1)
	cur := first.AddDate(0, 0, -int(first.Weekday()))

	var weeks [][]Day
	for !cur.After(last) {
		week := make([]Day, 7)
		for i := range week {
			week[i] = Day{Date: cur, InMonth: cur.Month() == month}
			cur = cur.AddDate(0, 0, 1)
		}
		weeks = append(weeks, week)
	}
	return weeks
}

func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// WorkingDays counts Monday-to-Friday days in the month; holidays are ignored.
func WorkingDays(year int, month time.Month) int {
	n := 0
	for d := 1; d <= DaysInMonth(year, month); d++ {
		switch time.Date(year, month, d, 0, 0, 0, 0, time.UTC).Weekday() {
		case time.Saturday, time.Sunday:
		default:
			n++
		}
	}
	return n
}
