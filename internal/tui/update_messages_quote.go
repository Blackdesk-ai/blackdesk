package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) handleFundamentalsLoaded(msg fundamentalsLoadedMsg) (Model, tea.Cmd) {
	if msg.err == nil {
		m.fundamentals = msg.data
		m.cacheFundamentals(msg.data)
		m.errFundamentals = nil
		m.profileScroll = 0
		return m, nil
	}
	if cached, ok := m.cachedFundamentals(msg.symbol); ok {
		m.fundamentals = cached
		m.profileScroll = 0
	} else {
		m.fundamentals = msg.data
	}
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

func (m Model) handleOwnersLoaded(msg ownersLoadedMsg) (Model, tea.Cmd) {
	m.owners = msg.data
	if msg.err == nil {
		m.cacheOwners(msg.data)
		items := m.ownerItemsForSymbol(msg.data.Symbol)
		if len(items) == 0 || m.ownersSel >= len(items) {
			m.ownersSel = 0
		}
	}
	m.errOwners = msg.err
	return m, nil
}

func (m Model) handleAnalystRecommendationsLoaded(msg analystRecommendationsLoadedMsg) (Model, tea.Cmd) {
	m.analyst = msg.data
	if msg.err == nil {
		m.cacheAnalystRecommendations(msg.data)
		items := m.analystRecommendationsForSymbol(msg.data.Symbol).Items
		if len(items) == 0 || m.analystSel >= len(items) {
			m.analystSel = 0
		}
	}
	m.errAnalyst = msg.err
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

func (m Model) handleEarningsLoaded(msg earningsLoadedMsg) (Model, tea.Cmd) {
	m.earnings = msg.data
	if msg.err == nil {
		m.cacheEarnings(msg.data)
		items := m.earningsItemsForSymbol(msg.data.Symbol)
		if len(items) == 0 || m.earningsSel >= len(items) {
			m.earningsSel = 0
		}
	}
	m.errEarnings = msg.err
	return m, nil
}
