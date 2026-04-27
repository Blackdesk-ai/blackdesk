package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
)

func (m Model) handleGlobalPageKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.globalPageOpen {
		return m, nil, false
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit, true
	case "ctrl+backspace", "ctrl+h":
		if next, cmd, handled := m.restorePreviousNavigation(); handled {
			return next, cmd, true
		}
		return m, nil, true
	case "ctrl+k":
		return m, m.openCommandPalette(), true
	case "?":
		m.helpOpen = true
		return m, nil, true
	case "esc":
		m.globalPageOpen = false
		m.status = "Closed calendar"
		return m, nil, true
	case "up":
		if m.globalPageKind == globalPageCalendar {
			m.cycleCalendarSelection(-1)
			return m, nil, true
		}
	case "down":
		if m.globalPageKind == globalPageCalendar {
			m.cycleCalendarSelection(1)
			return m, nil, true
		}
	case "left":
		if m.globalPageKind == globalPageCalendar {
			return m.changeCalendarFilter(-1)
		}
	case "right":
		if m.globalPageKind == globalPageCalendar {
			return m.changeCalendarFilter(1)
		}
	case "r":
		if m.globalPageKind == globalPageCalendar {
			delete(m.calendarCache, m.calendarFilter)
			m.calendar = domain.EconomicCalendarSnapshot{}
			m.errCalendar = nil
			m.calendarSel = 0
			m.status = "Refreshing economic calendar…"
			return m, m.loadCalendarCmd(m.calendarFilter), true
		}
	}
	return m, nil, true
}

func (m Model) changeCalendarFilter(step int) (Model, tea.Cmd, bool) {
	next := m.calendarFilter
	switch next {
	case calendarFilterToday:
		if step > 0 {
			next = calendarFilterThisWeek
		}
	case calendarFilterThisWeek:
		if step < 0 {
			next = calendarFilterToday
		}
	}
	if next == m.calendarFilter {
		return m, nil, true
	}
	m.calendarFilter = next
	m.calendarSel = 0
	m.status = "Calendar filter: " + m.calendarFilterLabel()
	if data, ok := m.cachedCalendar(next); ok {
		m.calendar = data
		m.errCalendar = nil
		return m, nil, true
	}
	m.calendar = domain.EconomicCalendarSnapshot{}
	m.errCalendar = nil
	return m, m.loadCalendarCmd(next), true
}
