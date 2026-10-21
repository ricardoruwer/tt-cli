package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/ricardoruwer/tt-cli/internal/model"
)

func sampleModel(w, h int) Model {
	now := time.Date(2026, time.October, 21, 10, 0, 0, 0, time.Local)
	m := New(nil, now)
	m.width, m.height = w, h
	m.loading = false
	m.byDay = model.IndexByDay([]model.Entry{
		{Date: "2026-10-09", Time: "08:00:00", Description: "- Created Story VIRT-3720\n- [VIRT-3716] Investigate the alert for high cost of Datadog logs\n- [VIRT-3713] Restock parent ATP once on auth-ship CancelAndRestockItem\n- Meeting: Ricardo / Diego\n- [VIRT-3700, VIRT-3698] Code Review\n- [VIRT-3707] Deploy", CodeReviews: "VIRT-3700, VIRT-3698", Deploys: "VIRT-3707"},
		{Date: "2026-10-21", Time: "08:00:00", Description: "- [VIRT-3729] Skip no-op grouping representative recompute\n- [VIRT-3751] Code Review"},
	}, 2026, time.October)
	return m
}

func TestGridFillsTerminalExactly(t *testing.T) {
	for _, size := range [][2]int{{140, 40}, {100, 30}, {80, 24}} {
		m := sampleModel(size[0], size[1])
		lines := strings.Split(m.render(), "\n")
		if len(lines) != size[1] {
			t.Errorf("%dx%d: got %d lines", size[0], size[1], len(lines))
		}
		// Grid body lines (after the title and weekday header) span the full width.
		for i := 2; i < len(lines)-2; i++ {
			if w := ansi.StringWidth(lines[i]); w != size[0] {
				t.Errorf("%dx%d: line %d width = %d", size[0], size[1], i, w)
			}
		}
	}
}

func TestDetailShowsFullDescription(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 9
	m.detail = true
	out := ansi.Strip(m.render())
	if !strings.Contains(out, "Friday, October 9, 2026") || !strings.Contains(out, "[VIRT-3707] Deploy") {
		t.Errorf("detail view missing content:\n%s", out)
	}
}

func TestDetailShowsCodeReviewsAndDeploys(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 9
	m.detail = true
	out := ansi.Strip(m.render())
	for _, want := range []string{"Code Reviews", "VIRT-3698", "Deploys"} {
		if !strings.Contains(out, want) {
			t.Errorf("detail view missing %q:\n%s", want, out)
		}
	}

	m.selected = 21
	if out := ansi.Strip(m.render()); strings.Contains(out, "Deploys") {
		t.Errorf("day without deploys should not show the section:\n%s", out)
	}
}

func TestTinyTerminalShowsMessage(t *testing.T) {
	if out := sampleModel(40, 10).render(); !strings.Contains(out, "Terminal too small") {
		t.Errorf("expected too-small message, got %q", out)
	}
}

func TestPrintGrid(t *testing.T) {
	if testing.Verbose() {
		t.Log("\n" + ansi.Strip(sampleModel(140, 36).render()))
	}
}
