package tui

type globalPageKind int

const (
	globalPageCalendar globalPageKind = iota
)

type calendarFilterMode int

const (
	calendarFilterToday calendarFilterMode = iota
	calendarFilterThisWeek
)
