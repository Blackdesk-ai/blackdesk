package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleWatchlistSelectionDebounced(msg watchlistSelectionDebouncedMsg) (Model, tea.Cmd) {
	if msg.id != m.watchlistSelectionDebounceID {
		return m, nil
	}
	if m.tabIdx != tabQuote {
		return m, nil
	}
	if msg.symbol == "" || msg.symbol != m.activeSymbol() {
		return m, nil
	}
	return m, m.loadAllCmd(msg.symbol)
}
