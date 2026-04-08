package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m Model) cachedInsiders(symbol string) (domain.InsiderSnapshot, bool) {
	if strings.TrimSpace(symbol) == "" {
		return domain.InsiderSnapshot{}, false
	}
	if strings.EqualFold(m.insiders.Symbol, symbol) {
		return m.insiders, true
	}
	if m.insiderCache == nil {
		return domain.InsiderSnapshot{}, false
	}
	data, ok := m.insiderCache[strings.ToUpper(strings.TrimSpace(symbol))]
	if !ok {
		return domain.InsiderSnapshot{}, false
	}
	return data, true
}

func (m *Model) cacheInsiders(data domain.InsiderSnapshot) {
	if strings.TrimSpace(data.Symbol) == "" {
		return
	}
	if m.insiderCache == nil {
		m.insiderCache = make(map[string]domain.InsiderSnapshot)
	}
	m.insiderCache[strings.ToUpper(strings.TrimSpace(data.Symbol))] = data
}

func (m Model) needsInsiders(symbol string) bool {
	if strings.TrimSpace(symbol) == "" {
		return false
	}
	if !m.services.HasInsiders() {
		return false
	}
	data, ok := m.cachedInsiders(symbol)
	if !ok {
		return true
	}
	return strings.TrimSpace(data.Symbol) == ""
}
