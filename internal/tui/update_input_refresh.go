package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleManualRefreshKey() (Model, tea.Cmd) {
	if next, cmd, handled := m.handleAIWorkspaceActionKey("r"); handled {
		return next, cmd
	}
	if m.tabIdx == tabScreener {
		plan := application.PlanManualRefresh(application.ManualRefreshInput{
			Workspace:         application.WorkspaceScreener,
			ScreenerAvailable: m.screenerAvailable(),
		})
		m.status = plan.Status
		if plan.RefreshScreener {
			return m, m.loadScreenerCmd(true)
		}
		return m, nil
	}
	workspace := application.WorkspaceQuote
	if m.tabIdx == tabNews {
		workspace = application.WorkspaceNews
	} else if m.tabIdx == tabMarkets {
		workspace = application.WorkspaceMarkets
	}
	plan := application.PlanManualRefresh(application.ManualRefreshInput{
		Workspace:         workspace,
		ActiveSymbol:      m.activeSymbol(),
		ScreenerAvailable: m.screenerAvailable(),
	})
	m.status = plan.Status
	if plan.TouchNewsClock {
		m.lastMarketNews = time.Now()
	}
	cmds := make([]tea.Cmd, 0, 2)
	if plan.RefreshAll {
		cmds = append(cmds, m.loadAllCmd(m.activeSymbol()))
		if m.tabIdx == tabQuote && m.quoteCenterMode == quoteCenterFilings {
			cmds = append(cmds, m.loadFilingsCmd(m.activeSymbol()))
		}
	}
	if plan.RefreshScreener {
		cmds = append(cmds, m.loadScreenerCmd(true))
	}
	if plan.RefreshMarketNews {
		cmds = append(cmds, m.loadMarketNewsCmd())
	}
	if plan.RefreshMarketSnap {
		cmds = append(cmds, m.loadMarketQuotesCmd())
	}
	return m, tea.Batch(cmds...)
}
