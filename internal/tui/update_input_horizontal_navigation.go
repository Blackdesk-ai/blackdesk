package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleTimeframeNavigation(step int) (Model, tea.Cmd, bool) {
	if !m.canChangeTimeframe() {
		return m, nil, true
	}
	currentIdx := m.rangeIdx
	if m.quoteCenterMode == quoteCenterSharpe {
		currentIdx = m.sharpeRangeIdx
	}
	plan := application.PlanWrappedIndexStep(application.WrappedIndexInput{
		CurrentIndex: currentIdx,
		Count:        len(ranges),
		Step:         step,
	})
	if !plan.Changed {
		return m, nil, true
	}
	if m.quoteCenterMode == quoteCenterSharpe {
		m.sharpeRangeIdx = plan.NextIndex
		m.status = "Risk Adjusted timeframe: " + ranges[m.sharpeRangeIdx].Label
		return m, nil, true
	}
	m.rangeIdx = plan.NextIndex
	m.updateRangeDefaults()
	m.touchAIContext()
	return m, tea.Batch(m.persistCmd(), m.loadHistoryCmd(m.activeSymbol())), true
}

func (m Model) handleStatisticsRangeNavigation(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterStatistics {
		return m, nil, false
	}
	plan := application.PlanWrappedIndexStep(application.WrappedIndexInput{
		CurrentIndex: m.statisticsRangeIdx,
		Count:        len(statisticsRangeSpecs),
		Step:         step,
	})
	if !plan.Changed {
		return m, nil, true
	}
	m.statisticsRangeIdx = plan.NextIndex
	m.status = "Statistics range: " + m.statisticsRangeSpec().Label
	return m, m.loadStatisticsHistoryCmd(m.activeSymbol()), true
}

func (m Model) handleFilingsFilterNavigation(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterFilings {
		return m, nil, false
	}
	m.cycleFilingsFilter(step)
	m.status = "Filings filter: " + m.filingsFilterLabel()
	return m, nil, true
}

func (m Model) handleStatementKindNavigation(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterStatements {
		return m, nil, false
	}
	if !m.services.HasStatements() {
		return m, nil, true
	}
	plan := application.StepStatementKind(m.statementKind, step)
	if m.statementKind != plan.Kind {
		m.statementKind = plan.Kind
		m.touchAIContext()
	}
	m.status = plan.Status
	return m, m.loadStatementCmd(m.activeSymbol()), true
}

func (m Model) handleStatementFrequencyNavigation(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterStatements {
		return m, nil, false
	}
	if !m.services.HasStatements() {
		return m, nil, true
	}
	plan := application.StepStatementFrequency(m.statementFreq, step)
	if m.statementFreq != plan.Frequency {
		m.statementFreq = plan.Frequency
		m.touchAIContext()
	}
	m.status = plan.Status
	return m, m.loadStatementCmd(m.activeSymbol()), true
}

func (m Model) handleScreenerDefinitionKey(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabScreener {
		return m, nil, false
	}
	if !m.screenerAvailable() {
		return m, nil, true
	}
	plan := application.PlanScreenerAdvance(application.ScreenerAdvanceInput{
		Available:    true,
		Definitions:  m.screenerDefs,
		CurrentIndex: m.screenerIdx,
		Step:         step,
	})
	m.screenerIdx = plan.NextIndex
	if plan.ApplyStatus {
		m.status = plan.Status
	}
	if plan.ShouldLoad {
		return m, m.loadScreenerCmd(true), true
	}
	return m, nil, true
}
