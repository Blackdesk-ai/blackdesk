package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m *Model) cacheEarnings(snapshot domain.EarningsSnapshot) {
	key := strings.ToUpper(strings.TrimSpace(snapshot.Symbol))
	if key == "" {
		return
	}
	m.earningsCache[key] = snapshot
}

func (m Model) cachedEarnings(symbol string) (domain.EarningsSnapshot, bool) {
	key := strings.ToUpper(strings.TrimSpace(symbol))
	if key == "" {
		return domain.EarningsSnapshot{}, false
	}
	if snapshot, ok := m.earningsCache[key]; ok {
		return snapshot, true
	}
	if strings.EqualFold(m.earnings.Symbol, key) {
		return m.earnings, true
	}
	return domain.EarningsSnapshot{}, false
}

func (m Model) earningsForSymbol(symbol string) domain.EarningsSnapshot {
	snapshot, _ := m.cachedEarnings(symbol)
	return snapshot
}

func (m Model) earningsItemsForSymbol(symbol string) []domain.EarningsItem {
	return m.earningsForSymbol(symbol).Items
}

func (m *Model) cycleEarningsSelection(step int) {
	items := m.earningsItemsForSymbol(m.activeSymbol())
	if len(items) == 0 {
		m.earningsSel = 0
		return
	}
	m.earningsSel = (m.earningsSel + step + len(items)) % len(items)
}

func (m Model) currentEarningsItem() (domain.EarningsItem, bool) {
	items := m.earningsItemsForSymbol(m.activeSymbol())
	if len(items) == 0 || m.earningsSel < 0 || m.earningsSel >= len(items) {
		return domain.EarningsItem{}, false
	}
	return items[m.earningsSel], true
}
