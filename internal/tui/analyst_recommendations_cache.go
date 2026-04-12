package tui

import (
	"strings"

	"blackdesk/internal/domain"
)

func (m *Model) cacheAnalystRecommendations(snapshot domain.AnalystRecommendationsSnapshot) {
	key := strings.ToUpper(strings.TrimSpace(snapshot.Symbol))
	if key == "" {
		return
	}
	m.analystCache[key] = snapshot
}

func (m Model) cachedAnalystRecommendations(symbol string) (domain.AnalystRecommendationsSnapshot, bool) {
	key := strings.ToUpper(strings.TrimSpace(symbol))
	if key == "" {
		return domain.AnalystRecommendationsSnapshot{}, false
	}
	if snapshot, ok := m.analystCache[key]; ok {
		return snapshot, true
	}
	if strings.EqualFold(m.analyst.Symbol, key) {
		return m.analyst, true
	}
	return domain.AnalystRecommendationsSnapshot{}, false
}

func (m Model) analystRecommendationsForSymbol(symbol string) domain.AnalystRecommendationsSnapshot {
	snapshot, _ := m.cachedAnalystRecommendations(symbol)
	return snapshot
}

func (m *Model) cycleAnalystRecommendationSelection(step int) {
	items := m.analystRecommendationsForSymbol(m.activeSymbol()).Items
	if len(items) == 0 {
		m.analystSel = 0
		return
	}
	m.analystSel = (m.analystSel + step + len(items)) % len(items)
}

func (m Model) currentAnalystRecommendation() (domain.AnalystRecommendationItem, bool) {
	items := m.analystRecommendationsForSymbol(m.activeSymbol()).Items
	if len(items) == 0 || m.analystSel < 0 || m.analystSel >= len(items) {
		return domain.AnalystRecommendationItem{}, false
	}
	return items[m.analystSel], true
}
