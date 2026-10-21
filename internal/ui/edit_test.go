package ui

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ricardoruwer/tt-cli/internal/model"
)

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "tab", "shift+tab", "enter", "esc", "backspace", "left", "right", "up", "down", "ctrl+s":
			msg = tea.KeyPressMsg(keyFor(k))
		default:
			msg = tea.KeyPressMsg(tea.Key{Code: []rune(k)[0], Text: k})
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func keyFor(name string) tea.Key {
	switch name {
	case "tab":
		return tea.Key{Code: tea.KeyTab}
	case "shift+tab":
		return tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}
	case "enter":
		return tea.Key{Code: tea.KeyEnter}
	case "esc":
		return tea.Key{Code: tea.KeyEscape}
	case "backspace":
		return tea.Key{Code: tea.KeyBackspace}
	case "left":
		return tea.Key{Code: tea.KeyLeft}
	case "right":
		return tea.Key{Code: tea.KeyRight}
	case "up":
		return tea.Key{Code: tea.KeyUp}
	case "down":
		return tea.Key{Code: tea.KeyDown}
	}
	return tea.Key{Code: 's', Mod: tea.ModCtrl}
}

func TestEditPrefillsFromEntry(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 9
	m = press(m, "e")
	if !m.editing {
		t.Fatal("e should start editing")
	}
	u := m.edit.update()
	if u.Time != "08:00" || u.Deploys != "VIRT-3707" || u.CodeReviews != "VIRT-3700, VIRT-3698" {
		t.Errorf("prefill = %+v", u)
	}
	if !strings.HasPrefix(u.Description, "- Created Story VIRT-3720") {
		t.Errorf("description = %q", u.Description)
	}
}

func TestEditTypingAndNavigation(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 12                                // no entry yet
	m = press(m, "e", "9", ":", "3", "0", "enter") // time, then enter moves on
	if m.edit.focus != fieldDescription {
		t.Fatalf("focus = %d", m.edit.focus)
	}
	m = press(m, "h", "i", "enter", "y", "o", "up", "x")
	if got := m.edit.fields[fieldDescription].String(); got != "hix\nyo" {
		t.Errorf("description = %q", got)
	}
	m = press(m, "backspace", "tab", "V", "-", "1")
	u := m.edit.update()
	if u.Time != "9:30" || u.Description != "hi\nyo" || u.CodeReviews != "V-1" {
		t.Errorf("update = %+v", u)
	}
}

func TestEditEscCancelsWithoutChangingEntry(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 9
	m = press(m, "e", "z", "esc")
	if m.editing {
		t.Error("esc should leave edit mode")
	}
	if strings.Contains(m.byDay[9].Time, "z") {
		t.Error("entry changed")
	}
}

func TestSavedMsgUpdatesEntryAndClosesEditor(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 12
	m = press(m, "e")
	m.edit.saving = true
	next, _ := m.Update(savedMsg{year: m.year, month: m.month, day: 12,
		entry: model.Entry{Date: "2026-10-12", Time: "07:30:00", Description: "- new"}})
	m = next.(Model)
	if m.editing || m.byDay[12].Minutes() != 450 {
		t.Errorf("editing=%v entry=%+v", m.editing, m.byDay[12])
	}
}

func TestSaveErrorKeepsEditorOpen(t *testing.T) {
	m := sampleModel(100, 30)
	m = press(m, "e")
	m.edit.saving = true
	next, _ := m.Update(savedMsg{year: m.year, month: m.month, day: m.selected, err: errors.New("boom")})
	m = next.(Model)
	if !m.editing || m.edit.saving || m.edit.err != "boom" {
		t.Errorf("editing=%v saving=%v err=%q", m.editing, m.edit.saving, m.edit.err)
	}
	if out := ansi.Strip(m.render()); !strings.Contains(out, "Could not save: boom") {
		t.Errorf("error not shown:\n%s", out)
	}
}

func TestEditRenderFillsTerminal(t *testing.T) {
	for _, size := range [][2]int{{140, 40}, {100, 30}, {80, 24}, {70, 19}} {
		m := sampleModel(size[0], size[1])
		m.selected = 9
		m = press(m, "e")
		lines := strings.Split(m.render(), "\n")
		if len(lines) != size[1] {
			t.Errorf("%dx%d: got %d lines", size[0], size[1], len(lines))
		}
		for i, l := range lines {
			if w := ansi.StringWidth(l); w > size[0] {
				t.Errorf("%dx%d: line %d width %d", size[0], size[1], i, w)
			}
		}
	}
}

func TestPasteIsSanitised(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 12
	m = press(m, "e")
	next, _ := m.Update(tea.PasteMsg{Content: "a\nb"})
	m = next.(Model)
	if got := m.edit.fields[fieldTime].String(); got != "" {
		t.Errorf("time field = %q", got)
	}
	m = press(m, "tab", "tab")
	next, _ = m.Update(tea.PasteMsg{Content: "a\nb"})
	m = next.(Model)
	if got := m.edit.fields[fieldCodeReviews].String(); got != "a b" {
		t.Errorf("code reviews field = %q", got)
	}
}

func TestTimeFieldOnlyAcceptsDigitsAndColon(t *testing.T) {
	m := sampleModel(100, 30)
	m.selected = 12
	m = press(m, "e", "a", "8", " ", ":", "-", "3", "0")
	if got := m.edit.fields[fieldTime].String(); got != "8:30" {
		t.Errorf("time field = %q", got)
	}
	next, _ := m.Update(tea.PasteMsg{Content: " 1x2 "})
	m = next.(Model)
	if got := m.edit.fields[fieldTime].String(); got != "8:3012" {
		t.Errorf("after paste = %q", got)
	}
}
