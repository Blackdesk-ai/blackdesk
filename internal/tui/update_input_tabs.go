package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleGlobalTopLevelKey(key string) (Model, tea.Cmd, bool) {
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit, true
	case "?":
		m.helpOpen = true
		return m, nil, true
	case "ctrl+k":
		return m, m.openCommandPalette(), true
	case "/":
		m.searchMode = true
		m.searchInput.SetValue("")
		m.searchItems = nil
		m.searchIdx = 0
		m.helpOpen = false
		m.searchInput.Focus()
		return m, nil, true
	case "c":
		if m.tabIdx != tabAI {
			return m, nil, false
		}
		m.searchMode = false
		m.aiFocused = false
		m.aiInput.Blur()
		m.aiPickerOpen = true
		m.aiPickerStep = aiPickerStepConnector
		m.aiModelBusy = false
		m.aiModelErr = nil
		m.status = "Select AI connector"
		return m, nil, true
	case "tab":
		return m, m.setActiveTab((m.tabIdx + 1) % len(headerTabs)), true
	case "1", "2", "3", "4", "5":
		return m, m.setActiveTab(int(key[0] - '1')), true
	case "enter":
		if m.tabIdx == tabQuote && m.quoteCenterMode == quoteCenterFilings {
			if item, ok := m.currentFiling(); ok && item.URL != "" {
				_ = openURLFunc(item.URL)
				m.status = "Opened SEC filing in browser"
			}
			return m, nil, true
		}
		if m.tabIdx != tabScreener {
			return m, nil, true
		}
		item, ok := m.currentScreenerItem()
		plan := application.PlanScreenerSymbolAction(application.ScreenerSymbolActionInput{
			Action:  application.ScreenerSymbolOpenQuote,
			HasItem: ok,
			Symbol:  item.Symbol,
		})
		if !plan.ApplyStatus {
			return m, nil, true
		}
		if plan.AddWatchlist {
			m.addToWatchlist(item.Symbol)
		}
		if plan.SelectSymbol {
			m.selectSymbol(item.Symbol)
		}
		if plan.OpenQuote {
			m.tabIdx = tabQuote
		}
		m.status = plan.Status
		return m, tea.Batch(m.persistCmd(), m.loadAllCmd(item.Symbol)), true
	default:
		return m, nil, false
	}
}

func (m Model) handleGlobalWorkspaceActionKey(key string) (Model, tea.Cmd, bool) {
	if next, cmd, handled := m.handleAIWorkspaceActionKey(key); handled {
		return next, cmd, true
	}
	if next, cmd, handled := m.handleQuoteWorkspaceActionKey(key); handled {
		return next, cmd, true
	}
	if next, cmd, handled := m.handleMarketsWorkspaceActionKey(key); handled {
		return next, cmd, true
	}
	if next, cmd, handled := m.handleScreenerWorkspaceActionKey(key); handled {
		return next, cmd, true
	}
	if next, cmd, handled := m.handleNewsWorkspaceActionKey(key); handled {
		return next, cmd, true
	}
	switch key {
	case "r":
		next, cmd := m.handleManualRefreshKey()
		return next, cmd, true
	case "u":
		if !m.updateAvailable || strings.TrimSpace(m.latestVersion) == "" || m.upgradeRunning {
			return m, nil, m.upgradeRunning
		}
		m.upgradeRunning = true
		m.status = "Upgrading Blackdesk…"
		return m, m.upgradeCmd(), true
	default:
		return m, nil, false
	}
}
