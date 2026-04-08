package tui

import "strings"

func (m Model) statusMetaText() string {
	provider := "unknown"
	if name := strings.TrimSpace(m.services.ActiveProviderName()); name != "" {
		switch strings.ToLower(name) {
		case "yahoo":
			provider = "Yahoo Finance"
		default:
			provider = name
		}
	}
	ai := m.activeAIStatusLabel()
	return "Source: " + provider + " | AI: " + ai
}

func (m Model) hasSufficientMarketOpinionData() bool {
	loaded := 0
	for _, symbol := range marketDashboardSymbols {
		if _, ok := m.watchQuotes[strings.ToUpper(symbol)]; ok {
			loaded++
		}
	}
	return loaded >= 12
}
