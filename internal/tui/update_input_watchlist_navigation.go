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
		cmds := []tea.Cmd{m.persistCmd(), m.loadAllCmd(m.activeSymbol())}
		if m.quoteCenterMode == quoteCenterFilings {
			cmds = append(cmds, m.loadFilingsCmd(m.activeSymbol()))
		}
		return m, tea.Batch(cmds...), true
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
		if m.tabIdx == tabQuote && m.quoteCenterMode == quoteCenterFilings {
			m.cycleFilingsSelection(step)
			return m, nil, true
		}
		return m.handleWatchlistNavigation(step)
	}
}
