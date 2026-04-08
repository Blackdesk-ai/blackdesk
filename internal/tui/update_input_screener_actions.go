package tui

import tea "github.com/charmbracelet/bubbletea"

import "blackdesk/internal/application"

func (m Model) handleScreenerWorkspaceActionKey(key string) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabScreener {
		return m, nil, false
	}
	switch key {
	case "a":
		item, ok := m.currentScreenerItem()
		plan := application.PlanScreenerSymbolAction(application.ScreenerSymbolActionInput{
			Action:  application.ScreenerSymbolAddWatch,
			HasItem: ok,
			Symbol:  item.Symbol,
		})
		if !plan.ApplyStatus {
			return m, nil, true
		}
		if plan.AddWatchlist {
			m.addToWatchlist(item.Symbol)
		}
		m.status = plan.Status
		return m, m.persistCmd(), true
	case "n":
		return m.handleScreenerDefinitionKey(1)
	case "p":
		return m.handleScreenerDefinitionKey(-1)
	default:
		return m, nil, false
	}
}
