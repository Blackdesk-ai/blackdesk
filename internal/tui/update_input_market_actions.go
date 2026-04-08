package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleMarketsWorkspaceActionKey(key string) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabMarkets {
		return m, nil, false
	}
	switch key {
	case "i":
		if m.hasSufficientMarketOpinionData() && !m.aiMarketOpinionRunning {
			m.aiMarketOpinionErr = nil
			m.pendingMarketOpinionRefresh = true
			m.status = "Refreshing market AI opinion…"
			return m, m.loadAllCmd(m.activeSymbol()), true
		}
		return m, nil, true
	default:
		return m, nil, false
	}
}
