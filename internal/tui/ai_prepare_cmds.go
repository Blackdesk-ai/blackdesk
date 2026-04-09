package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) prepareAIContextCmd(prompt string) tea.Cmd {
	symbol := m.activeSymbol()
	req := m.buildPrepareAIContextRequest(symbol, m.missingAIQuoteSymbols(symbol))

	return func() tea.Msg {
		result := m.services.PrepareAIContext(m.ctx, req)
		risk, riskErr := m.fetchMarketRiskSnapshot()
		msg := aiContextPreparedMsgFromResult(prompt, symbol, result)
		msg.marketRisk = risk
		msg.marketRiskErr = riskErr
		return msg
	}
}

func (m Model) prepareQuoteInsightCmd(symbol string) tea.Cmd {
	req := m.buildPrepareAIContextRequest(symbol, nil)

	return func() tea.Msg {
		result := m.services.PrepareAIContext(m.ctx, req)
		risk, riskErr := m.fetchMarketRiskSnapshot()
		msg := aiQuoteInsightPreparedMsgFromResult(symbol, result)
		msg.marketRisk = risk
		msg.marketRiskErr = riskErr
		return msg
	}
}
