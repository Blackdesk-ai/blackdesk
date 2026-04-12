package tui

import (
	"strings"
	"unicode"

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
		m.searchRequestQuery = ""
		m.searchInput.Blur()
		return m, nil, true
	case "enter":
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			return m, nil, true
		}
		if len(m.searchItems) > 0 {
			symbol := m.searchItems[m.searchIdx].Symbol
			next, cmd := m.openSearchSymbol(symbol)
			return next, cmd, true
		}
		if directSymbol, ok := normalizeDirectSearchSymbol(query); ok {
			next, cmd := m.openSearchSymbol(directSymbol)
			return next, cmd, true
		}
		if query == m.searchRequestQuery {
			m.status = "Searching…"
			return m, nil, true
		}
		m.status = "Searching…"
		m.searchRequestID++
		m.searchRequestQuery = query
		return m, m.searchCmd(query, m.searchRequestID), true
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
			query := strings.TrimSpace(m.searchInput.Value())
			m.searchDebounceID++
			m.searchRequestQuery = ""
			if query == "" {
				return m, cmd, true
			}
			return m, tea.Batch(cmd, m.searchDebounceCmd(query, m.searchDebounceID)), true
		}
		return m, cmd, true
	}
}

func (m Model) openSearchSymbol(symbol string) (Model, tea.Cmd) {
	m.addToWatchlist(symbol)
	m.selectSymbol(symbol)
	m.searchMode = false
	m.searchInput.Blur()
	m.searchRequestQuery = ""
	m.status = "Selected " + symbol
	return m, tea.Batch(m.persistCmd(), m.loadAllCmd(symbol))
}

func normalizeDirectSearchSymbol(query string) (string, bool) {
	symbol := strings.ToUpper(strings.TrimSpace(query))
	if symbol == "" || strings.ContainsAny(symbol, " \t\r\n") {
		return "", false
	}
	if strings.ContainsAny(symbol, ".-^=") {
		return symbol, true
	}
	if len(symbol) == 0 || len(symbol) > 4 {
		return "", false
	}
	for _, r := range symbol {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", false
		}
	}
	return symbol, true
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
