package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m *Model) setActiveTab(tab int) tea.Cmd {
	if tab < 0 {
		tab = 0
	}
	if tab >= len(headerTabs) {
		tab = len(headerTabs) - 1
	}
	prev := m.tabIdx
	prevMode := m.quoteCenterMode
	m.tabIdx = tab
	if m.tabIdx != prev || m.quoteCenterMode != prevMode {
		m.touchAIContext()
	}
	switch m.tabIdx {
	case tabNews:
		if prev == tabNews {
			return nil
		}
		m.lastMarketNews = time.Now()
		m.status = "Loading market news…"
		return tea.Batch(m.loadMarketNewsCmd(), m.loadMarketQuotesCmd())
	case tabScreener:
		plan := application.PlanScreenerEntry(application.ScreenerEntryInput{
			Available:   m.screenerAvailable(),
			WasActive:   prev == tabScreener,
			HasItems:    len(m.screenerResult.Items) > 0,
			CurrentName: m.currentScreenerDefinition().Name,
		})
		if plan.MarkLoaded {
			m.screenerLoaded = true
		}
		if plan.ApplyStatus {
			m.status = plan.Status
		}
		if plan.ShouldLoad {
			return m.loadScreenerCmd(true)
		}
		return nil
	default:
		return nil
	}
}

func (m Model) canChangeTimeframe() bool {
	return m.tabIdx == tabQuote && m.quoteCenterMode == quoteCenterChart
}

func (m *Model) setQuoteCenterMode(mode quoteCenterMode) {
	if m.quoteCenterMode != mode {
		m.quoteCenterMode = mode
		m.touchAIContext()
	}
}

func (m Model) quoteBottomPanelsVisible() bool {
	return m.tabIdx == tabQuote &&
		m.quoteCenterMode != quoteCenterStatements &&
		m.quoteCenterMode != quoteCenterInsiders &&
		m.quoteCenterMode != quoteCenterAnalyst &&
		m.quoteCenterMode != quoteCenterFilings &&
		m.quoteCenterMode != quoteCenterEarnings
}

func (m *Model) activeSymbol() string {
	if len(m.config.Watchlist) == 0 {
		return "AAPL"
	}
	if m.selectedIdx < 0 || m.selectedIdx >= len(m.config.Watchlist) {
		m.selectedIdx = 0
	}
	return m.config.Watchlist[m.selectedIdx]
}

func (m *Model) updateRangeDefaults() {
	current := ranges[m.rangeIdx]
	m.config = application.SetDefaultRange(m.config, current.Range, current.Interval)
}

func (m Model) applicationQuoteCenterMode() application.QuoteCenterMode {
	switch m.quoteCenterMode {
	case quoteCenterTechnicals:
		return application.QuoteCenterTechnicals
	case quoteCenterStatements:
		return application.QuoteCenterStatements
	case quoteCenterInsiders:
		return application.QuoteCenterInsiders
	case quoteCenterAnalyst:
		return application.QuoteCenterAnalyst
	case quoteCenterFilings:
		return application.QuoteCenterFilings
	case quoteCenterEarnings:
		return application.QuoteCenterEarnings
	case quoteCenterFundamentals:
		return application.QuoteCenterFundamentals
	case quoteCenterNews:
		return application.QuoteCenterNews
	default:
		return application.QuoteCenterChart
	}
}
