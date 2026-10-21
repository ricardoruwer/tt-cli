package ui

import "testing"

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
