package ui

import (
	"testing"
	"time"
)

func TestSelectionSkipsWeekends(t *testing.T) {
	m := sampleModel(100, 30) // October 2026: the 9th is a Friday
	m.selected = 9

	m.stepWeekday(1)
	if m.selected != 12 {
		t.Errorf("l from Friday: got %d, want Monday 12", m.selected)
	}
	m.stepWeekday(-1)
	if m.selected != 9 {
		t.Errorf("h from Monday: got %d, want Friday 9", m.selected)
	}

	m.move(7)
	if m.selected != 16 {
		t.Errorf("j from Friday 9: got %d, want 16", m.selected)
	}
}

func TestSelectionStaysInsideMonth(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 1 // Thursday
	m.stepWeekday(-1)
	m.move(-7)
	if m.selected != 1 {
		t.Errorf("got %d, want 1", m.selected)
	}
	m.selected = 30 // Friday
	m.stepWeekday(1)
	m.move(7)
	if m.selected != 30 {
		t.Errorf("got %d, want 30", m.selected)
	}
}

func TestNearestWeekday(t *testing.T) {
	m := sampleModel(100, 30)
	for day, want := range map[int]int{10: 9, 11: 9, 12: 12, 3: 2} {
		if got := m.nearestWeekday(day); got != want {
			t.Errorf("nearestWeekday(%d) = %d, want %d", day, got, want)
		}
	}
}

func TestShiftMonthWrapsYearsAndReloads(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 21

	m, cmd := m.shiftMonth(3) // January 2027
	if m.year != 2027 || m.month != time.January || cmd == nil || !m.loading {
		t.Fatalf("got %d-%d loading=%v cmd=%v", m.year, m.month, m.loading, cmd != nil)
	}
	if m.selected != 21 || len(m.byDay) != 0 {
		t.Errorf("selected=%d entries=%d", m.selected, len(m.byDay))
	}

	m, _ = m.shiftMonth(-3)
	if m.year != 2026 || m.month != time.October {
		t.Errorf("got %d-%d, want 2026-10", m.year, m.month)
	}
}

func TestShiftMonthClampsDay(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 30
	m, _ = m.shiftMonth(4) // February 2027 has 28 days; the 28th is a Sunday
	if m.selected != 26 {
		t.Errorf("selected = %d, want Friday 26", m.selected)
	}
}

func TestTodayReturnsToCurrentMonth(t *testing.T) {
	m := sampleModel(100, 30)
	m, _ = m.shiftMonth(2)
	next, cmd := m.handleKey("t")
	m = next.(Model)
	if m.month != time.October || m.year != 2026 || m.selected != 21 || cmd == nil {
		t.Errorf("got %d-%d day %d", m.year, m.month, m.selected)
	}
}
