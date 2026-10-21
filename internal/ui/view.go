package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ricardoruwer/tt-cli/internal/calendar"
	"github.com/ricardoruwer/tt-cli/internal/model"
)

var (
	colBgIn      = lipgloss.Color("#111827")
	colBgOut     = lipgloss.Color("#1c2333")
	colBgSel     = lipgloss.Color("#26335c")
	colBgHeader  = lipgloss.Color("#161c2b")
	colRule      = lipgloss.Color("#2a3140")
	colText      = lipgloss.Color("#d1d5db")
	colDim       = lipgloss.Color("#6b7280")
	colBadgeBg   = lipgloss.Color("#3730a3")
	colBadgeText = lipgloss.Color("#e0e7ff")
	colToday     = lipgloss.Color("#6366f1")
	colError     = lipgloss.Color("#f87171")
)

var weekdayNames = []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}

const (
	minCellWidth  = 10
	minCellHeight = 3
	chromeLines   = 4 // title, weekday header, stats, help
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	weeks := calendar.Weeks(m.year, m.month)
	needW, needH := 7*minCellWidth, chromeLines+len(weeks)*minCellHeight
	if m.width < needW || m.height < needH {
		return fmt.Sprintf("Terminal too small: need at least %dx%d, have %dx%d",
			needW, needH, m.width, m.height)
	}

	var body string
	switch {
	case m.detail:
		body = m.renderDetail()
	case m.err != nil:
		body = m.renderError()
	default:
		body = m.renderGrid(weeks)
	}

	lines := []string{m.renderTitle()}
	if !m.detail {
		lines = append(lines, m.renderWeekdays())
	} else {
		lines = append(lines, "")
	}
	lines = append(lines, body, m.renderStats(), m.renderHelp())
	return strings.Join(lines, "\n")
}

func (m Model) bodyHeight() int { return m.height - chromeLines }

func (m Model) renderTitle() string {
	title := lipgloss.NewStyle().Bold(true).Render(m.month.String()) +
		" " + lipgloss.NewStyle().Foreground(colDim).Render(fmt.Sprint(m.year))
	if m.loading {
		title += lipgloss.NewStyle().Foreground(colDim).Render("  loading…")
	}
	return title
}

func (m Model) renderWeekdays() string {
	widths := columnWidths(m.width)
	cells := make([]string, 7)
	for i, name := range weekdayNames {
		cells[i] = lipgloss.NewStyle().
			Background(colBgHeader).Foreground(colDim).
			Width(widths[i]).Align(lipgloss.Center).
			Render(name)
	}
	return strings.Join(cells, "")
}

func (m Model) renderGrid(weeks [][]calendar.Day) string {
	widths := columnWidths(m.width)
	avail := m.bodyHeight()
	base, rem := avail/len(weeks), avail%len(weeks)

	var out []string
	for r, week := range weeks {
		h := base
		if r < rem {
			h++
		}
		cells := make([][]string, 7)
		for c, day := range week {
			cells[c] = m.renderCell(day, widths[c], h)
		}
		for i := 0; i < h; i++ {
			var sb strings.Builder
			for c := range cells {
				sb.WriteString(cells[c][i])
			}
			out = append(out, sb.String())
		}
	}
	return strings.Join(out, "\n")
}

// renderCell returns exactly h lines, each w columns wide. The last column of
// every line and the last line of every cell form the grid rules.
func (m Model) renderCell(day calendar.Day, w, h int) []string {
	inner := w - 1
	bg := colBgIn
	if !day.InMonth {
		bg = colBgOut
	}
	selected := day.InMonth && day.Date.Day() == m.selected
	if selected {
		bg = colBgSel
	}
	base := lipgloss.NewStyle().Background(bg)
	ruleCol := lipgloss.NewStyle().Background(colRule).Render(" ")

	finish := func(s string) string {
		pad := inner - ansi.StringWidth(s)
		if pad > 0 {
			s += base.Render(strings.Repeat(" ", pad))
		}
		return s + ruleCol
	}

	// Day number on the left, time badge on the right.
	numStyle := base.Foreground(colDim)
	if day.InMonth {
		numStyle = base.Foreground(colText)
	}
	if day.InMonth && m.isToday(day.Date) {
		numStyle = lipgloss.NewStyle().Background(colToday).Foreground(lipgloss.Color("#ffffff")).Bold(true)
	}
	if selected {
		numStyle = numStyle.Bold(true)
	}
	num := base.Render(" ") + numStyle.Render(fmt.Sprintf(" %02d ", day.Date.Day()))

	entry, hasEntry := model.Entry{}, false
	if day.InMonth {
		entry, hasEntry = m.byDay[day.Date.Day()]
	}

	head := num
	if hasEntry {
		text := model.FormatHM(entry.Minutes())
		if inner >= 14 {
			text = " " + text + " "
		}
		badge := lipgloss.NewStyle().Background(colBadgeBg).Foreground(colBadgeText).Bold(true).Render(text)
		gap := inner - ansi.StringWidth(num) - ansi.StringWidth(badge)
		if gap < 0 {
			gap = 0
		}
		head = num + base.Render(strings.Repeat(" ", gap)) + badge
	}

	lines := []string{finish(head)}

	if hasEntry {
		descLines := entry.Lines()
		avail := h - 2
		if avail > 0 && len(descLines) > avail {
			hidden := len(descLines) - (avail - 1)
			descLines = append(descLines[:avail-1], fmt.Sprintf("+%d more", hidden))
		}
		textStyle := base.Foreground(colText)
		for i := 0; i < len(descLines) && len(lines) < h-1; i++ {
			t := ansi.Truncate(strings.TrimPrefix(descLines[i], "- "), max(inner-2, 1), "…")
			lines = append(lines, finish(base.Render(" ")+textStyle.Render(t)))
		}
	}

	for len(lines) < h-1 {
		lines = append(lines, finish(""))
	}
	lines = append(lines, lipgloss.NewStyle().Background(colRule).Render(strings.Repeat(" ", w)))
	return lines
}

func (m Model) isToday(d time.Time) bool {
	return d.Year() == m.today.Year() && d.Month() == m.today.Month() && d.Day() == m.today.Day()
}

func (m Model) renderError() string {
	msg := lipgloss.NewStyle().Foreground(colError).Bold(true).Render("Could not load time entries") +
		"\n\n" + m.err.Error() +
		"\n\n" + lipgloss.NewStyle().Foreground(colDim).Render("press r to retry, q to quit")
	return lipgloss.Place(m.width, m.bodyHeight(), lipgloss.Center, lipgloss.Center, msg)
}

func (m Model) renderStats() string {
	total := model.TotalMinutes(m.byDay)
	chip := func(label, value string) string {
		return lipgloss.NewStyle().Foreground(colDim).Render(label+" ") +
			lipgloss.NewStyle().Bold(true).Render(value)
	}
	return chip("Current working time", model.FormatHM(total)) + "   " +
		chip("Total working days", fmt.Sprintf("%d days", calendar.WorkingDays(m.year, m.month)))
}

func (m Model) renderHelp() string {
	help := "←↓↑→/hjkl move · enter details · t today · r reload · q quit"
	if m.detail {
		help = "j/k scroll · space/pgdn page · g/G top/bottom · esc back · ctrl+c quit"
	}
	return lipgloss.NewStyle().Foreground(colDim).Render(help)
}

// Detail panel.

func (m Model) detailVisible() int { return max(m.bodyHeight()-2, 1) }

func (m Model) detailLines() []string {
	entry, ok := m.byDay[m.selected]
	if !ok {
		return []string{"No entries for this day."}
	}
	wrap := lipgloss.NewStyle().Width(max(m.width-4, 10))
	heading := lipgloss.NewStyle().Bold(true).Foreground(colDim)

	var out []string
	add := func(lines []string) {
		for _, l := range lines {
			out = append(out, strings.Split(wrap.Render(l), "\n")...)
		}
	}
	section := func(title string, items []string) {
		if len(items) == 0 {
			return
		}
		if len(out) > 0 {
			out = append(out, "")
		}
		out = append(out, heading.Render(title))
		add(items)
	}

	section("Description", entry.Lines())
	section("Code Reviews", nonEmpty(entry.CodeReviews))
	section("Deploys", nonEmpty(entry.Deploys))
	if len(out) == 0 {
		return []string{"No details for this day."}
	}
	return out
}

func (m Model) renderDetail() string {
	date := time.Date(m.year, m.month, m.selected, 0, 0, 0, 0, time.UTC)
	title := lipgloss.NewStyle().Bold(true).Render(date.Format("Monday, January 2, 2006"))
	if e, ok := m.byDay[m.selected]; ok {
		title += "  " + lipgloss.NewStyle().Background(colBadgeBg).Foreground(colBadgeText).Bold(true).
			Render(" "+model.FormatHM(e.Minutes())+" ")
	}

	lines := m.detailLines()
	visible := m.detailVisible()
	start := min(max(m.scroll, 0), max(len(lines)-visible, 0))
	end := min(start+visible, len(lines))
	window := lines[start:end]

	out := []string{title, ""}
	for _, l := range window {
		out = append(out, "  "+l)
	}
	for len(out) < m.bodyHeight() {
		out = append(out, "")
	}
	return strings.Join(out, "\n")
}

// nonEmpty wraps a trimmed value as a single line, or nothing when it is blank.
func nonEmpty(s string) []string {
	if s = strings.TrimSpace(s); s == "" {
		return nil
	}
	return []string{s}
}

// columnWidths splits the terminal width across 7 columns, giving the
// remainder to the leftmost ones so the grid always fills the screen.
func columnWidths(total int) []int {
	base, rem := total/7, total%7
	widths := make([]int, 7)
	for i := range widths {
		widths[i] = base
		if i < rem {
			widths[i]++
		}
	}
	return widths
}
