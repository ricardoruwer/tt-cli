package model

import (
	"reflect"
	"testing"
	"time"
)

func TestMinutes(t *testing.T) {
	cases := map[string]int{"08:00:00": 480, "00:30:00": 30, "": 0, "garbage": 0}
	for in, want := range cases {
		if got := (Entry{Time: in}).Minutes(); got != want {
			t.Errorf("Minutes(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestFormatHM(t *testing.T) {
	cases := map[int]string{0: "00:00", 480: "08:00", 7200: "120:00", 90: "01:30"}
	for in, want := range cases {
		if got := FormatHM(in); got != want {
			t.Errorf("FormatHM(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestLines(t *testing.T) {
	e := Entry{Description: "- a\n\n  - b  \n"}
	want := []string{"- a", "- b"}
	if got := e.Lines(); !reflect.DeepEqual(got, want) {
		t.Errorf("Lines() = %v, want %v", got, want)
	}
}

func TestIndexByDayIgnoresOrderAndOtherMonths(t *testing.T) {
	entries := []Entry{
		{Date: "2026-10-09", Time: "08:00:00"},
		{Date: "2026-10-01", Time: "04:00:00"},
		{Date: "2026-09-30", Time: "08:00:00"},
		{Date: "bad", Time: "08:00:00"},
	}
	byDay := IndexByDay(entries, 2026, time.October)
	if len(byDay) != 2 || byDay[1].Time != "04:00:00" || byDay[9].Time != "08:00:00" {
		t.Fatalf("unexpected index: %v", byDay)
	}
	if got := TotalMinutes(byDay); got != 720 {
		t.Errorf("TotalMinutes = %d, want 720", got)
	}
}
