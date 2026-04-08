package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleWatchlistNavigation(step int) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote {
		return m, nil, true
	}
	plan := application.PlanWatchlistNavigation(application.WatchlistNavigationInput{
		CurrentIndex: m.selectedIdx,
		Count:        len(m.config.Watchlist),
		Step:         step,
	})
	if plan.Changed {
		m.selectedIdx = plan.NextIndex
		m.ensureWatchlistSelectionVisible()
		m.selectSymbol(m.activeSymbol())
		return m, tea.Batch(m.persistCmd(), m.loadAllCmd(m.activeSymbol())), true
	}
	return m, nil, true
}

func (m Model) handleWorkspaceVerticalNavigation(step int) (Model, tea.Cmd, bool) {
	switch m.tabIdx {
	case tabAI:
		m.scrollAITranscript(-step)
		return m, nil, true
	case tabScreener:
		m.cycleScreenerSelection(step)
		return m, nil, true
	case tabNews:
		m.cycleMarketNewsSelection(step)
		return m, nil, true
	default:
		return m.handleWatchlistNavigation(step)
	}
}
