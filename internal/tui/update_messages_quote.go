package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleFundamentalsLoaded(msg fundamentalsLoadedMsg) (Model, tea.Cmd) {
	m.fundamentals = msg.data
	m.errFundamentals = msg.err
	m.profileScroll = 0
	return m, nil
}

func (m Model) handleStatementLoaded(msg statementLoadedMsg) (Model, tea.Cmd) {
	m.statement = msg.data
	if msg.err == nil {
		m.cacheStatement(msg.data)
	}
	m.errStatement = msg.err
	return m, nil
}

func (m Model) handleInsidersLoaded(msg insidersLoadedMsg) (Model, tea.Cmd) {
	m.insiders = msg.data
	if msg.err == nil {
		m.cacheInsiders(msg.data)
	}
	m.errInsiders = msg.err
	return m, nil
}

func (m Model) handleFilingsLoaded(msg filingsLoadedMsg) (Model, tea.Cmd) {
	m.filings = msg.data
	if msg.err == nil {
		m.cacheFilings(msg.data)
		items := m.filteredFilingsSnapshot(msg.data.Symbol).Items
		if len(items) == 0 {
			m.filingsSel = 0
		} else if m.filingsSel >= len(items) {
			m.filingsSel = 0
		}
	}
	m.errFilings = msg.err
	return m, nil
}
