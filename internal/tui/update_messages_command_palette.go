package tui

import "strings"

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleCommandPaletteDebounced(msg commandPaletteDebouncedMsg) (Model, tea.Cmd) {
	if !m.commandPaletteOpen || msg.id != m.commandPaletteDebounceID {
		return m, nil
	}
	query := strings.TrimSpace(m.commandInput.Value())
	if query == "" || query != msg.query || query == m.commandPaletteRequestQuery {
		return m, nil
	}
	m.status = "Searching symbols…"
	m.commandPaletteRequestID++
	m.commandPaletteRequestQuery = query
	return m, m.commandPaletteSearchCmd(query, m.commandPaletteRequestID)
}

func (m Model) handleCommandPaletteLoaded(msg commandPaletteLoadedMsg) (Model, tea.Cmd) {
	if !m.commandPaletteOpen || msg.id != m.commandPaletteRequestID || msg.query != m.commandPaletteRequestQuery {
		return m, nil
	}
	if msg.err != nil {
		m.status = msg.err.Error()
		m.commandPaletteSymbolItems = nil
		m.refreshCommandPaletteItems()
		return m, nil
	}
	m.commandPaletteSymbolItems = msg.results
	m.refreshCommandPaletteItems()
	return m, nil
}
