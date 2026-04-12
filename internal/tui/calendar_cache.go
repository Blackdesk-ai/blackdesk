package tui

import (
	"time"

	"blackdesk/internal/domain"
)

func (m Model) cachedCalendar(filter calendarFilterMode) (domain.EconomicCalendarSnapshot, bool) {
	start, end := m.calendarRangeForFilter(filter)
	if !m.calendar.StartDate.IsZero() && m.calendarFilter == filter && sameCalendarRange(m.calendar, start, end) {
		return m.calendar, true
	}
	data, ok := m.calendarCache[filter]
	if !ok || !sameCalendarRange(data, start, end) {
		return domain.EconomicCalendarSnapshot{}, false
	}
	return data, true
}

func (m *Model) cacheCalendar(filter calendarFilterMode, data domain.EconomicCalendarSnapshot) {
	if m.calendarCache == nil {
		m.calendarCache = make(map[calendarFilterMode]domain.EconomicCalendarSnapshot)
	}
	m.calendarCache[filter] = data
}

func (m *Model) cycleCalendarSelection(step int) {
	items := m.calendar.Events
	if len(items) == 0 || step == 0 {
		m.calendarSel = 0
		return
	}
	m.calendarSel = clamp(m.calendarSel+step, 0, len(items)-1)
}

func (m Model) currentCalendarEvent() (domain.EconomicCalendarEvent, bool) {
	if m.calendarSel < 0 || m.calendarSel >= len(m.calendar.Events) {
		return domain.EconomicCalendarEvent{}, false
	}
	return m.calendar.Events[m.calendarSel], true
}

func sameCalendarRange(snapshot domain.EconomicCalendarSnapshot, start, end time.Time) bool {
	return snapshot.StartDate.Equal(start) && snapshot.EndDate.Equal(end)
}
