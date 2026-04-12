package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleAIResponseLoaded(msg aiResponseLoadedMsg) (Model, tea.Cmd) {
	m.aiRunning = false
	m.aiDuration = msg.duration
	m.aiErr = msg.err
	m.aiOutput = strings.TrimSpace(msg.output)
	m.aiLastRequestTruncation = msg.truncation
	if msg.contextSent != "" {
		m.aiLastContext = msg.contextSent
		m.aiLastSymbol = msg.symbol
		m.aiLastContextRevision = msg.contextRevision
	}
	m.pushAIAssistantMessage(m.aiOutput, msg.err, msg.duration)
	if msg.err != nil {
		m.status = fmt.Sprintf("%s failed", m.activeAIConnectorLabel())
	} else {
		m.status = fmt.Sprintf("%s replied in %s", m.activeAIConnectorLabel(), msg.duration.Round(time.Millisecond))
		if msg.truncation.hasAny() {
			m.status += " • clipped request"
		}
	}
	return m, nil
}

func (m Model) handleAIMarketOpinionLoaded(msg aiMarketOpinionLoadedMsg) (Model, tea.Cmd) {
	m.aiMarketOpinionRunning = false
	m.pendingMarketOpinionRefresh = false
	m.aiMarketOpinionErr = msg.err
	m.aiMarketOpinion = compactAIInsight(msg.output, 220)
	for symbol, series := range msg.histories {
		key := strings.ToUpper(symbol)
		m.marketOpinionHistory[key] = series
		m.marketOpinionHistoryAt[key] = time.Now()
	}
	if msg.err == nil && m.aiMarketOpinion != "" {
		m.aiMarketOpinionUpdated = time.Now()
		m.status = fmt.Sprintf("Market AI opinion updated in %s", msg.duration.Round(time.Millisecond))
	} else if msg.err != nil {
		m.status = fmt.Sprintf("%s market opinion failed", m.activeAIConnectorLabel())
	}
	return m, nil
}

func (m Model) handleAIQuoteInsightLoaded(msg aiQuoteInsightLoadedMsg) (Model, tea.Cmd) {
	m.aiQuoteInsightRunning = false
	m.aiQuoteInsightSymbol = strings.ToUpper(msg.symbol)
	m.aiQuoteInsightErr = msg.err
	m.aiQuoteInsight = compactAIInsight(msg.output, 180)
	if msg.err == nil && m.aiQuoteInsight != "" {
		m.aiQuoteInsightUpdated = time.Now()
		m.status = fmt.Sprintf("Quote AI insight updated in %s", msg.duration.Round(time.Millisecond))
	} else if msg.err != nil {
		m.status = fmt.Sprintf("%s quote insight failed", m.activeAIConnectorLabel())
	}
	return m, nil
}

func (m Model) handleAIModelsLoaded(msg aiModelsLoadedMsg) (Model, tea.Cmd) {
	m.aiModelBusy = false
	m.aiModelErr = msg.err
	if msg.err == nil {
		m.aiModels[strings.ToLower(msg.connectorID)] = append([]string(nil), msg.models...)
		m.syncAIModelSelection()
		if len(msg.models) > 0 {
			m.status = fmt.Sprintf("Loaded %d models for %s", len(msg.models), m.activeAIConnectorLabel())
		} else {
			m.status = fmt.Sprintf("No models reported by %s", m.activeAIConnectorLabel())
		}
	} else {
		m.status = msg.err.Error()
	}
	return m, nil
}
