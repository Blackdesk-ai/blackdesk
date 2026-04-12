package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleCalendarLoaded(msg calendarLoadedMsg) (Model, tea.Cmd) {
	if msg.err == nil {
		m.cacheCalendar(msg.filter, msg.data)
		m.calendar = msg.data
		if len(m.calendar.Events) == 0 || m.calendarSel >= len(m.calendar.Events) {
			m.calendarSel = 0
		}
	}
	m.errCalendar = msg.err
	return m, nil
}
