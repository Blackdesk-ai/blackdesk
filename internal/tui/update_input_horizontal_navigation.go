package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleTimeframeNavigation(step int) (Model, tea.Cmd, bool) {
	if !m.canChangeTimeframe() {
		return m, nil, true
	}
	plan := application.PlanWrappedIndexStep(application.WrappedIndexInput{
		CurrentIndex: m.rangeIdx,
		Count:        len(ranges),
		Step:         step,
	})
	if !plan.Changed {
		return m, nil, true
	}
	m.rangeIdx = plan.NextIndex
	m.updateRangeDefaults()
	return m, tea.Batch(m.persistCmd(), m.loadHistoryCmd(m.activeSymbol())), true
}

func (m Model) handleStatementKindNavigation(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterStatements {
		return m, nil, false
	}
	if !m.services.HasStatements() {
		return m, nil, true
	}
	plan := application.StepStatementKind(m.statementKind, step)
	m.statementKind = plan.Kind
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
	m.statementFreq = plan.Frequency
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
