package ui

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// editFixedLines is every body line except the description text: title, blank,
// four labels, three single-line inputs, four blanks and the status line.
const editFixedLines = 14

func (m Model) renderEdit() string {
	width := max(m.width-4, 10)
	descRows := max(m.bodyHeight()-editFixedLines, 1)

	date := time.Date(m.year, m.month, m.selected, 0, 0, 0, 0, time.UTC)
	out := []string{
		lipgloss.NewStyle().Bold(true).Render("Edit " + date.Format("Monday, January 2, 2006")),
		"",
	}

	for i := range m.edit.fields {
		rows := 1
		if i == fieldDescription {
			rows = descRows
		}
		focused := i == m.edit.focus
		label := lipgloss.NewStyle().Bold(true).Foreground(colDim)
		if focused {
			label = lipgloss.NewStyle().Bold(true).Foreground(colToday)
		}
		out = append(out, label.Render(fieldLabels[i]))
		for _, l := range m.edit.fields[i].render(width, rows, focused) {
			out = append(out, "  "+l)
		}
		out = append(out, "")
	}

	out = append(out, m.renderEditStatus())

	for len(out) < m.bodyHeight() {
		out = append(out, "")
	}
	return strings.Join(out[:m.bodyHeight()], "\n")
}

func (m Model) renderEditStatus() string {
	switch {
	case m.edit.saving:
		return lipgloss.NewStyle().Foreground(colDim).Render("saving…")
	case m.edit.err != "":
		return lipgloss.NewStyle().Foreground(colError).Bold(true).Render("Could not save: " + m.edit.err)
	}
	return ""
}

// render returns exactly rows lines of at most width columns. The window
// follows the cursor line, and the line holding the cursor scrolls
// horizontally instead of being truncated.
func (f field) render(width, rows int, focused bool) []string {
	lines := strings.Split(string(f.text), "\n")

	curLine, curCol := 0, f.cur
	for i, l := range lines {
		n := len([]rune(l))
		if curCol <= n {
			curLine = i
			break
		}
		curCol -= n + 1
	}

	first := min(max(curLine-rows+1, 0), max(len(lines)-rows, 0))
	out := make([]string, 0, rows)
	for i := first; i < len(lines) && len(out) < rows; i++ {
		r := []rune(lines[i])
		if !focused || i != curLine {
			out = append(out, ansi.Truncate(string(r), width, "…"))
			continue
		}
		start := max(curCol-(width-1), 0)
		cursor, after := " ", ""
		if curCol < len(r) {
			cursor = string(r[curCol])
			after = string(r[curCol+1 : min(start+width, len(r))])
		}
		out = append(out, string(r[start:curCol])+
			lipgloss.NewStyle().Reverse(true).Render(cursor)+after)
	}
	for len(out) < rows {
		out = append(out, "")
	}
	return out
}
