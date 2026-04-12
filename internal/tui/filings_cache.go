package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m *Model) cacheFilings(snapshot domain.FilingsSnapshot) {
	key := strings.ToUpper(strings.TrimSpace(snapshot.Symbol))
	if key == "" {
		return
	}
	m.filingsCache[key] = snapshot
}

func (m Model) filingsForSymbol(symbol string) domain.FilingsSnapshot {
	key := strings.ToUpper(strings.TrimSpace(symbol))
	if key == "" {
		return domain.FilingsSnapshot{}
	}
	if snapshot, ok := m.filingsCache[key]; ok {
		return snapshot
	}
	if strings.EqualFold(m.filings.Symbol, key) {
		return m.filings
	}
	return domain.FilingsSnapshot{}
}

func (m *Model) cycleFilingsSelection(step int) {
	items := m.filingsForSymbol(m.activeSymbol()).Items
	if len(items) == 0 {
		m.filingsSel = 0
		return
	}
	m.filingsSel = (m.filingsSel + step + len(items)) % len(items)
}

func (m Model) currentFiling() (domain.FilingItem, bool) {
	items := m.filingsForSymbol(m.activeSymbol()).Items
	if len(items) == 0 || m.filingsSel < 0 || m.filingsSel >= len(items) {
		return domain.FilingItem{}, false
	}
	return items[m.filingsSel], true
}
