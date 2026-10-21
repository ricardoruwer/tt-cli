package calendar

import (
	"testing"
	"time"
)

func TestWeeksOctober2026(t *testing.T) {
	weeks := Weeks(2026, time.October)
	if len(weeks) != 5 {
		t.Fatalf("got %d weeks, want 5", len(weeks))
	}
	if d := weeks[0][0]; d.InMonth || d.Date.Day() != 27 || d.Date.Month() != time.September {
		t.Errorf("first cell = %v, want Sep 27 out of month", d)
	}
	if d := weeks[0][4]; !d.InMonth || d.Date.Day() != 1 {
		t.Errorf("Thursday of first row = %v, want Oct 1 in month", d)
	}
	if d := weeks[4][6]; !d.InMonth || d.Date.Day() != 31 {
		t.Errorf("last cell = %v, want Oct 31 in month", d)
	}
}

func TestWeeksRowCounts(t *testing.T) {
	// Feb 2026 starts on Sunday and has 28 days: exactly four rows.
	if n := len(Weeks(2026, time.February)); n != 4 {
		t.Errorf("Feb 2026 weeks = %d, want 4", n)
	}
	// Aug 2026 starts on Saturday and has 31 days: six rows.
	if n := len(Weeks(2026, time.August)); n != 6 {
		t.Errorf("Aug 2026 weeks = %d, want 6", n)
	}
}

func TestWorkingDays(t *testing.T) {
	if n := WorkingDays(2026, time.October); n != 22 {
		t.Errorf("Oct 2026 working days = %d, want 22", n)
	}
	if n := WorkingDays(2026, time.February); n != 20 {
		t.Errorf("Feb 2026 working days = %d, want 20", n)
	}
}
