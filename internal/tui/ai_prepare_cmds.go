package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
)

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

func (m Model) prepareFilingAnalysisCmd(symbol string, item domain.FilingItem) tea.Cmd {
	req := m.buildPrepareAIContextRequest(symbol, nil)

	return func() tea.Msg {
		result := m.services.PrepareAIContext(m.ctx, req)
		risk, riskErr := m.fetchMarketRiskSnapshot()
		filing, filingErr := m.services.GetFilingDocument(m.ctx, item)
		msg := aiFilingAnalysisPreparedMsgFromResult(filingAnalysisPrompt(symbol, item), symbol, filing, filingErr, result)
		msg.marketRisk = risk
		msg.marketRiskErr = riskErr
		return msg
	}
}
