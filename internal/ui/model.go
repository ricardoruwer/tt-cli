package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ricardoruwer/tt-cli/internal/api"
	"github.com/ricardoruwer/tt-cli/internal/calendar"
	"github.com/ricardoruwer/tt-cli/internal/model"
)

type loadedMsg struct {
	year    int
	month   time.Month
	entries []model.Entry
	err     error
}

type Model struct {
	client   *api.Client
	year     int
	month    time.Month
	today    time.Time
	selected int
	byDay    map[int]model.Entry
	loading  bool
	err      error

	width, height int

	detail bool
	scroll int
}

func New(client *api.Client, now time.Time) Model {
	m := Model{
		client:  client,
		year:    now.Year(),
		month:   now.Month(),
		today:   now,
		byDay:   map[int]model.Entry{},
		loading: true,
	}
	m.selected = m.nearestWeekday(now.Day())
	return m
}

func (m Model) Init() tea.Cmd {
	return m.fetch()
}

func (m Model) fetch() tea.Cmd {
	client, year, month := m.client, m.year, m.month
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		entries, err := client.Month(ctx, year, month)
		return loadedMsg{year: year, month: month, entries: entries, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampScroll()

	case loadedMsg:
		if msg.year != m.year || msg.month != m.month {
			return m, nil
		}
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.byDay = model.IndexByDay(msg.entries, m.year, m.month)
		}

	case tea.KeyPressMsg:
		return m.handleKey(msg.String())
	}
	return m, nil
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	if m.detail {
		switch key {
		case "esc", "enter", "q":
			m.detail, m.scroll = false, 0
		case "j", "down":
			m.scroll++
		case "k", "up":
			m.scroll--
		case "pgdown", "space":
			m.scroll += m.detailVisible()
		case "pgup":
			m.scroll -= m.detailVisible()
		case "g", "home":
			m.scroll = 0
		case "G", "end":
			m.scroll = 1 << 30
		}
		m.clampScroll()
		return m, nil
	}

	switch key {
	case "q":
		return m, tea.Quit
	case "h", "left":
		m.stepWeekday(-1)
	case "l", "right":
		m.stepWeekday(1)
	case "k", "up":
		m.move(-7)
	case "j", "down":
		m.move(7)
	case "enter":
		m.detail, m.scroll = true, 0
	case "t":
		if m.today.Year() == m.year && m.today.Month() == m.month {
			m.selected = m.nearestWeekday(m.today.Day())
		}
	case "r":
		m.loading, m.err = true, nil
		return m, m.fetch()
	}
	return m, nil
}

func (m Model) isWeekend(day int) bool {
	return calendar.IsWeekend(time.Date(m.year, m.month, day, 0, 0, 0, 0, time.UTC))
}

// nearestWeekday returns day, or the closest earlier weekday when it falls on a
// weekend.
func (m Model) nearestWeekday(day int) int {
	for day > 1 && m.isWeekend(day) {
		day--
	}
	if m.isWeekend(day) {
		return m.firstWeekday()
	}
	return day
}

func (m Model) firstWeekday() int {
	day := 1
	for m.isWeekend(day) {
		day++
	}
	return day
}

// stepWeekday moves the selection to the adjacent weekday, staying put at the
// edges of the month.
func (m *Model) stepWeekday(dir int) {
	last := calendar.DaysInMonth(m.year, m.month)
	for d := m.selected + dir; d >= 1 && d <= last; d += dir {
		if !m.isWeekend(d) {
			m.selected = d
			return
		}
	}
}

// move jumps by whole weeks, so the selection stays on the same weekday; it
// does nothing when the target is outside the month.
func (m *Model) move(delta int) {
	if d := m.selected + delta; d >= 1 && d <= calendar.DaysInMonth(m.year, m.month) {
		m.selected = d
	}
}

func (m *Model) clampScroll() {
	maxScroll := max(0, len(m.detailLines())-m.detailVisible())
	m.scroll = min(max(m.scroll, 0), maxScroll)
}
