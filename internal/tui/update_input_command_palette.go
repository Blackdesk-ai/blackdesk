package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleCommandPaletteKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.commandPaletteOpen {
		return m, nil, false
	}
	switch msg.String() {
	case "esc", "ctrl+k":
		m.closeCommandPalette("Closed command palette")
		return m, nil, true
	case "enter":
		next, cmd := m.executeCommandPaletteSelection()
		return next, cmd, true
	case "up":
		if m.commandPaletteIdx > 0 {
			m.commandPaletteIdx--
		}
		return m, nil, true
	case "down":
		if m.commandPaletteIdx < len(m.commandPaletteItems)-1 {
			m.commandPaletteIdx++
		}
		return m, nil, true
	default:
		var cmd tea.Cmd
		prevValue := m.commandInput.Value()
		m.commandInput, cmd = m.commandInput.Update(msg)
		if m.commandInput.Value() != prevValue {
			query := strings.TrimSpace(m.commandInput.Value())
			m.commandPaletteSymbolItems = nil
			m.commandPaletteDebounceID++
			m.commandPaletteRequestQuery = ""
			m.refreshCommandPaletteItems()
			if query == "" {
				return m, cmd, true
			}
			return m, tea.Batch(cmd, m.commandPaletteDebounceCmd(query, m.commandPaletteDebounceID)), true
		}
		return m, cmd, true
	}
}
