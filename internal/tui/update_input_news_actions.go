package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleNewsWorkspaceActionKey(key string) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabNews {
		return m, nil, false
	}
	switch key {
	case "n":
		m.cycleMarketNewsSelection(1)
		return m, nil, true
	case "p":
		m.cycleMarketNewsSelection(-1)
		return m, nil, true
	case "o":
		if len(m.marketNews) > 0 {
			_ = openURLFunc(m.marketNews[m.marketNewsSel].URL)
			m.status = "Opened market news story in browser"
		}
		return m, nil, true
	default:
		return m, nil, false
	}
}
