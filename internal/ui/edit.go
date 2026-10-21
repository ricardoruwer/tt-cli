package ui

import (
	"context"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ricardoruwer/tt-cli/internal/api"
	"github.com/ricardoruwer/tt-cli/internal/model"
)

const (
	fieldTime = iota
	fieldDescription
	fieldCodeReviews
	fieldDeploys
	fieldCount
)

var fieldLabels = [fieldCount]string{"Time", "Description", "Code Reviews", "Deploys"}

type savedMsg struct {
	year  int
	month time.Month
	day   int
	entry model.Entry
	err   error
}

// field is a text buffer with a cursor, both measured in runes. Only the
// description spans several lines.
type field struct {
	text []rune
	cur  int
}

func newField(s string) field {
	r := []rune(s)
	return field{text: r, cur: len(r)}
}

func (f *field) insert(s string) {
	r := []rune(s)
	f.text = append(f.text[:f.cur:f.cur], append(r, f.text[f.cur:]...)...)
	f.cur += len(r)
}

func (f *field) backspace() {
	if f.cur == 0 {
		return
	}
	f.text = append(f.text[:f.cur-1:f.cur-1], f.text[f.cur:]...)
	f.cur--
}

func (f *field) delete() {
	if f.cur >= len(f.text) {
		return
	}
	f.text = append(f.text[:f.cur:f.cur], f.text[f.cur+1:]...)
}

func (f *field) move(delta int) {
	f.cur = min(max(f.cur+delta, 0), len(f.text))
}

// lineStart and lineEnd bound the line the cursor is on.
func (f field) lineStart() int {
	i := f.cur
	for i > 0 && f.text[i-1] != '\n' {
		i--
	}
	return i
}

func (f field) lineEnd() int {
	i := f.cur
	for i < len(f.text) && f.text[i] != '\n' {
		i++
	}
	return i
}

// moveLine moves the cursor one line up or down, keeping the column when the
// target line is long enough. It reports false at the first/last line.
func (f *field) moveLine(dir int) bool {
	start := f.lineStart()
	col := f.cur - start
	if dir < 0 {
		if start == 0 {
			return false
		}
		prevEnd := start - 1
		g := field{text: f.text, cur: prevEnd}
		f.cur = min(g.lineStart()+col, prevEnd)
		return true
	}
	end := f.lineEnd()
	if end >= len(f.text) {
		return false
	}
	next := end + 1
	g := field{text: f.text, cur: next}
	f.cur = min(next+col, g.lineEnd())
	return true
}

func (f field) String() string { return string(f.text) }

type editor struct {
	fields [fieldCount]field
	focus  int
	saving bool
	err    string
}

func newEditor(e model.Entry) editor {
	return editor{fields: [fieldCount]field{
		fieldTime:        newField(trimSeconds(e.Time)),
		fieldDescription: newField(e.Description),
		fieldCodeReviews: newField(e.CodeReviews),
		fieldDeploys:     newField(e.Deploys),
	}}
}

// trimSeconds shortens the API's "HH:MM:SS" to "HH:MM" for editing.
func trimSeconds(t string) string {
	if strings.Count(t, ":") == 2 && strings.HasSuffix(t, ":00") {
		return strings.TrimSuffix(t, ":00")
	}
	return t
}

func (e editor) update() api.Update {
	return api.Update{
		Time:        strings.TrimSpace(e.fields[fieldTime].String()),
		Description: strings.TrimRight(e.fields[fieldDescription].String(), " \n"),
		CodeReviews: strings.TrimSpace(e.fields[fieldCodeReviews].String()),
		Deploys:     strings.TrimSpace(e.fields[fieldDeploys].String()),
	}
}

func (e *editor) cycle(dir int) {
	e.focus = (e.focus + dir + fieldCount) % fieldCount
}

func (m Model) startEdit() Model {
	m.editing = true
	m.edit = newEditor(m.byDay[m.selected])
	return m
}

func (m Model) save() tea.Cmd {
	client, year, month, day := m.client, m.year, m.month, m.selected
	u := m.edit.update()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		entry, err := client.Save(ctx, date, u)
		return savedMsg{year: year, month: month, day: day, entry: entry, err: err}
	}
}

func (m Model) handleEditKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.edit.saving {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	}

	f := &m.edit.fields[m.edit.focus]
	multiline := m.edit.focus == fieldDescription
	m.edit.err = ""

	key := msg.String()
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.editing = false
	case "ctrl+s":
		m.edit.saving = true
		return m, m.save()
	case "tab":
		m.edit.cycle(1)
	case "shift+tab":
		m.edit.cycle(-1)
	case "enter":
		if multiline {
			f.insert("\n")
		} else {
			m.edit.cycle(1)
		}
	case "backspace":
		f.backspace()
	case "delete":
		f.delete()
	case "left":
		f.move(-1)
	case "right":
		f.move(1)
	case "up":
		if !multiline || !f.moveLine(-1) {
			m.edit.cycle(-1)
		}
	case "down":
		if !multiline || !f.moveLine(1) {
			m.edit.cycle(1)
		}
	case "home", "ctrl+a":
		f.cur = f.lineStart()
	case "end", "ctrl+e":
		f.cur = f.lineEnd()
	default:
		if t := msg.Key().Text; t != "" {
			f.insert(sanitizeFor(m.edit.focus, t))
		}
	}
	return m, nil
}

func (m Model) handlePaste(text string) (tea.Model, tea.Cmd) {
	if !m.editing || m.edit.saving {
		return m, nil
	}
	m.edit.err = ""
	m.edit.fields[m.edit.focus].insert(sanitizeFor(m.edit.focus, text))
	return m, nil
}

// sanitizeFor filters typed or pasted text for a field. The time field keeps
// only digits and ':'.
func sanitizeFor(focus int, s string) string {
	if focus != fieldTime {
		return sanitize(s, focus == fieldDescription)
	}
	return strings.Map(func(r rune) rune {
		if r == ':' || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, s)
}

// sanitize normalises line endings and drops control characters; single-line
// fields also flatten newlines into spaces.
func sanitize(s string, multiline bool) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' && multiline:
			sb.WriteRune(r)
		case r == '\n' || r == '\t':
			sb.WriteRune(' ')
		case r >= ' ' && r != 0x7f:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
