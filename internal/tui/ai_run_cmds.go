package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/agents"
	"blackdesk/internal/domain"
)

func (m Model) runAICmd(prompt string) tea.Cmd {
	connectorID := m.activeAIConnectorID()
	if strings.TrimSpace(connectorID) == "" {
		return func() tea.Msg {
			return aiResponseLoadedMsg{connectorID: "", err: fmt.Errorf("no AI connector selected")}
		}
	}
	if m.services == nil {
		return func() tea.Msg {
			return aiResponseLoadedMsg{connectorID: connectorID, err: fmt.Errorf("AI registry unavailable")}
		}
	}
	envelope, err := m.buildAIRequest(strings.TrimSpace(prompt))
	if err != nil {
		return func() tea.Msg {
			return aiResponseLoadedMsg{connectorID: connectorID, err: err}
		}
	}
	request := agents.Request{
		Workspace:    m.workspaceRoot,
		Prompt:       envelope.Prompt,
		Model:        strings.TrimSpace(m.config.AIModel),
		SystemPrompt: envelope.SystemPrompt,
	}
	return func() tea.Msg {
		_ = writeAILastRequestDump(m.workspaceRoot, connectorID, request.Model, envelope)
		resp, err := m.services.RunAI(m.ctx, connectorID, request)
		return aiResponseLoadedMsg{
			connectorID:     connectorID,
			output:          resp.Output,
			duration:        resp.Duration,
			contextSent:     envelope.ContextPayload,
			contextRevision: envelope.ContextRevision,
			truncation:      envelope.Truncation,
			symbol:          envelope.ActiveSymbol,
			err:             err,
		}
	}
}

func (m Model) runMarketOpinionCmd() tea.Cmd {
	connectorID := m.activeAIConnectorID()
	if strings.TrimSpace(connectorID) == "" {
		return func() tea.Msg {
			return aiMarketOpinionLoadedMsg{connectorID: "", err: fmt.Errorf("no AI connector selected")}
		}
	}
	if m.services == nil {
		return func() tea.Msg {
			return aiMarketOpinionLoadedMsg{connectorID: connectorID, err: fmt.Errorf("AI registry unavailable")}
		}
	}
	request := agents.Request{
		Workspace: m.workspaceRoot,
		Model:     strings.TrimSpace(m.config.AIModel),
	}
	return func() tea.Msg {
		histories := make(map[string]domain.PriceSeries)
		for _, symbol := range marketOpinionHistorySymbols {
			key := strings.ToUpper(symbol)
			series, histErr := m.services.GetHistory(m.ctx, symbol, "1y", "1d")
			if histErr != nil {
				return aiMarketOpinionLoadedMsg{connectorID: connectorID, err: histErr}
			}
			histories[key] = series
		}
		if risk, riskErr := m.fetchMarketRiskSnapshot(); riskErr != nil {
			m.marketRisk = domain.MarketRiskSnapshot{}
		} else {
			m.marketRisk = risk
		}
		envelope, err := m.buildAIMarketOpinionRequest(histories)
		if err != nil {
			return aiMarketOpinionLoadedMsg{connectorID: connectorID, err: err}
		}
		request.SystemPrompt = envelope.SystemPrompt
		request.Prompt = envelope.Prompt
		resp, err := m.services.RunAI(m.ctx, connectorID, request)
		return aiMarketOpinionLoadedMsg{
			connectorID: connectorID,
			output:      resp.Output,
			duration:    resp.Duration,
			histories:   histories,
			err:         err,
		}
	}
}

func (m Model) runQuoteInsightCmd(symbol string) tea.Cmd {
	connectorID := m.activeAIConnectorID()
	if strings.TrimSpace(connectorID) == "" {
		return func() tea.Msg {
			return aiQuoteInsightLoadedMsg{connectorID: "", symbol: symbol, err: fmt.Errorf("no AI connector selected")}
		}
	}
	if m.services == nil {
		return func() tea.Msg {
			return aiQuoteInsightLoadedMsg{connectorID: connectorID, symbol: symbol, err: fmt.Errorf("AI registry unavailable")}
		}
	}
	envelope, err := m.buildAIQuoteInsightRequest(symbol)
	if err != nil {
		return func() tea.Msg {
			return aiQuoteInsightLoadedMsg{connectorID: connectorID, symbol: symbol, err: err}
		}
	}
	request := agents.Request{
		Workspace:    m.workspaceRoot,
		Prompt:       envelope.Prompt,
		Model:        strings.TrimSpace(m.config.AIModel),
		SystemPrompt: envelope.SystemPrompt,
	}
	return func() tea.Msg {
		_ = writeAILastRequestDump(m.workspaceRoot, connectorID, request.Model, envelope)
		resp, err := m.services.RunAI(m.ctx, connectorID, request)
		return aiQuoteInsightLoadedMsg{
			connectorID: connectorID,
			output:      resp.Output,
			duration:    resp.Duration,
			contextSent: envelope.ContextPayload,
			symbol:      symbol,
			err:         err,
		}
	}
}

func (m Model) runFilingAnalysisCmd(symbol string, snapshot domain.FilingsSnapshot, filing domain.FilingDocument, prompt string) tea.Cmd {
	connectorID := m.activeAIConnectorID()
	if strings.TrimSpace(connectorID) == "" {
		return func() tea.Msg {
			return aiResponseLoadedMsg{connectorID: "", symbol: symbol, err: fmt.Errorf("no AI connector selected")}
		}
	}
	if m.services == nil {
		return func() tea.Msg {
			return aiResponseLoadedMsg{connectorID: connectorID, symbol: symbol, err: fmt.Errorf("AI registry unavailable")}
		}
	}
	envelope, err := m.buildAIFilingAnalysisRequest(symbol, snapshot, filing, prompt)
	if err != nil {
		return func() tea.Msg {
			return aiResponseLoadedMsg{connectorID: connectorID, symbol: symbol, err: err}
		}
	}
	request := agents.Request{
		Workspace:    m.workspaceRoot,
		Prompt:       envelope.Prompt,
		Model:        strings.TrimSpace(m.config.AIModel),
		SystemPrompt: envelope.SystemPrompt,
	}
	return func() tea.Msg {
		_ = writeAILastRequestDump(m.workspaceRoot, connectorID, request.Model, envelope)
		resp, err := m.services.RunAI(m.ctx, connectorID, request)
		return aiResponseLoadedMsg{
			connectorID:     connectorID,
			output:          resp.Output,
			duration:        resp.Duration,
			contextSent:     envelope.ContextPayload,
			contextRevision: envelope.ContextRevision,
			truncation:      envelope.Truncation,
			symbol:          symbol,
			err:             err,
		}
	}
}
