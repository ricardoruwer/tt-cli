package model

import (
	"fmt"
	"strings"
	"time"
)

type Entry struct {
	Date        string `json:"date"`
	Time        string `json:"time"`
	Description string `json:"description"`
	// CodeReviews and Deploys are comma-separated ticket keys; the API sends
	// null when a day has none, which decodes to "".
	CodeReviews string `json:"code_reviews"`
	Deploys     string `json:"deploys"`
}

// HasTime reports whether a time was set; the API sends null for days without one.
func (e Entry) HasTime() bool { return strings.TrimSpace(e.Time) != "" }

// Minutes parses the "HH:MM:SS" time field; malformed values count as zero.
func (e Entry) Minutes() int {
	var h, m, s int
	if _, err := fmt.Sscanf(e.Time, "%d:%d:%d", &h, &m, &s); err != nil {
		return 0
	}
	return h*60 + m
}

// Lines returns the non-empty, trimmed lines of the description.
func (e Entry) Lines() []string {
	var lines []string
	for _, l := range strings.Split(e.Description, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// FormatHM renders minutes as H:MM with at least two hour digits ("08:00", "120:00").
func FormatHM(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

// IndexByDay keys the entries that belong to the given month by day of month,
// regardless of the order the API returned them in.
func IndexByDay(entries []Entry, year int, month time.Month) map[int]Entry {
	byDay := make(map[int]Entry, len(entries))
	for _, e := range entries {
		d, err := time.Parse("2006-01-02", e.Date)
		if err != nil || d.Year() != year || d.Month() != month {
			continue
		}
		byDay[d.Day()] = e
	}
	return byDay
}

func TotalMinutes(byDay map[int]Entry) int {
	total := 0
	for _, e := range byDay {
		total += e.Minutes()
	}
	return total
}
