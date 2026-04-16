package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m *Model) cacheFundamentals(snapshot domain.FundamentalsSnapshot) {
	key := strings.ToUpper(strings.TrimSpace(snapshot.Symbol))
	if key == "" {
		return
	}
	if m.fundamentalsCache == nil {
		m.fundamentalsCache = make(map[string]domain.FundamentalsSnapshot)
	}
	m.fundamentalsCache[key] = snapshot
}

func (m Model) cachedFundamentals(symbol string) (domain.FundamentalsSnapshot, bool) {
	key := strings.ToUpper(strings.TrimSpace(symbol))
	if key == "" {
		return domain.FundamentalsSnapshot{}, false
	}
	if snapshot, ok := m.fundamentalsCache[key]; ok {
		return snapshot, true
	}
	if strings.EqualFold(m.fundamentals.Symbol, key) {
		return m.fundamentals, true
	}
	return domain.FundamentalsSnapshot{}, false
}
