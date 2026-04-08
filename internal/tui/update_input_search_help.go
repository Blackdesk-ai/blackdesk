package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleSearchKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.searchMode {
		return m, nil, false
	}
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchInput.SetValue("")
		m.searchItems = nil
		m.searchIdx = 0
		m.searchInput.Blur()
		return m, nil, true
	case "enter":
		if len(m.searchItems) > 0 {
			symbol := m.searchItems[m.searchIdx].Symbol
			m.addToWatchlist(symbol)
			m.selectSymbol(symbol)
			m.searchMode = false
			m.searchInput.Blur()
			m.status = "Selected " + symbol
			return m, tea.Batch(m.persistCmd(), m.loadAllCmd(symbol)), true
		}
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			return m, nil, true
		}
		m.status = "Searching…"
		return m, m.searchCmd(query), true
	case "up":
		if m.searchIdx > 0 {
			m.searchIdx--
		}
		return m, nil, true
	case "down":
		if m.searchIdx < len(m.searchItems)-1 {
			m.searchIdx++
		}
		return m, nil, true
	case "ctrl+a":
		if len(m.searchItems) == 0 {
			return m, nil, true
		}
		symbol := m.searchItems[m.searchIdx].Symbol
		m.addToWatchlist(symbol)
		m.selectSymbol(symbol)
		return m, tea.Batch(m.persistCmd(), m.loadAllCmd(symbol)), true
	default:
		var cmd tea.Cmd
		prevValue := m.searchInput.Value()
		m.searchInput, cmd = m.searchInput.Update(msg)
		if m.searchInput.Value() != prevValue {
			m.searchItems = nil
			m.searchIdx = 0
		}
		return m, cmd, true
	}
}

func (m Model) handleHelpKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.helpOpen {
		return m, nil, false
	}
	switch msg.String() {
	case "esc", "?", "q":
		m.helpOpen = false
	}
	return m, nil, true
}
