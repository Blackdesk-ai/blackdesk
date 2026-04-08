package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) prepareAIContextCmd(prompt string) tea.Cmd {
	symbol := m.activeSymbol()
	req := m.buildPrepareAIContextRequest(symbol, m.missingAIQuoteSymbols(symbol))

	return func() tea.Msg {
		return aiContextPreparedMsgFromResult(prompt, symbol, m.services.PrepareAIContext(m.ctx, req))
	}
}

func (m Model) prepareQuoteInsightCmd(symbol string) tea.Cmd {
	req := m.buildPrepareAIContextRequest(symbol, nil)

	return func() tea.Msg {
		return aiQuoteInsightPreparedMsgFromResult(symbol, m.services.PrepareAIContext(m.ctx, req))
	}
}
