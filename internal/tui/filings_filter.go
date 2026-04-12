package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

type filingsFilterMode int

const (
	filingsFilterAll filingsFilterMode = iota
	filingsFilterTenK
	filingsFilterTenQ
	filingsFilterPeriodicReports
)

var filingsFilterModes = []filingsFilterMode{
	filingsFilterAll,
	filingsFilterPeriodicReports,
	filingsFilterTenK,
	filingsFilterTenQ,
}

func (m Model) filteredFilingsSnapshot(symbol string) domain.FilingsSnapshot {
	snapshot := m.filingsForSymbol(symbol)
	if len(snapshot.Items) == 0 {
		return snapshot
	}
	if m.filingsFilter == filingsFilterAll {
		return snapshot
	}
	filtered := make([]domain.FilingItem, 0, len(snapshot.Items))
	for _, item := range snapshot.Items {
		if m.filingMatchesFilter(item) {
			filtered = append(filtered, item)
		}
	}
	snapshot.Items = filtered
	return snapshot
}

func (m Model) filingMatchesFilter(item domain.FilingItem) bool {
	form := strings.ToUpper(strings.TrimSpace(item.Form))
	switch m.filingsFilter {
	case filingsFilterTenK:
		return form == "10-K"
	case filingsFilterTenQ:
		return form == "10-Q"
	case filingsFilterPeriodicReports:
		return form == "10-K" || form == "10-Q"
	default:
		return true
	}
}

func (m Model) filingsFilterLabel() string {
	switch m.filingsFilter {
	case filingsFilterTenK:
		return "10-K"
	case filingsFilterTenQ:
		return "10-Q"
	case filingsFilterPeriodicReports:
		return "10-K/10-Q"
	default:
		return "All"
	}
}

func (m *Model) cycleFilingsFilter(step int) {
	current := 0
	for i, mode := range filingsFilterModes {
		if mode == m.filingsFilter {
			current = i
			break
		}
	}
	next := (current + step + len(filingsFilterModes)) % len(filingsFilterModes)
	m.filingsFilter = filingsFilterModes[next]
	m.resetFilingsSelectionForFilter()
}

func (m *Model) resetFilingsSelectionForFilter() {
	items := m.filteredFilingsSnapshot(m.activeSymbol()).Items
	if len(items) == 0 {
		m.filingsSel = 0
		return
	}
	if m.filingsSel < 0 || m.filingsSel >= len(items) {
		m.filingsSel = 0
	}
}
