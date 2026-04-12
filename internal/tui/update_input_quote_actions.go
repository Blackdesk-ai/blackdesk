package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
)

func (m Model) handleQuoteWorkspaceActionKey(key string) (Model, tea.Cmd, bool) {
	if m.tabIdx != tabQuote {
		return m, nil, false
	}
	switch key {
	case "c":
		plan := application.PlanQuoteCenterSelection(application.QuoteCenterSelectionInput{
			Target: application.QuoteCenterChart,
		})
		if plan.Allowed {
			m.quoteCenterMode = quoteCenterChart
		}
		m.status = plan.Status
		return m, nil, true
	case "f":
		plan := application.PlanQuoteCenterSelection(application.QuoteCenterSelectionInput{
			Target: application.QuoteCenterFundamentals,
		})
		if plan.Allowed {
			m.quoteCenterMode = quoteCenterFundamentals
		}
		m.status = plan.Status
		return m, nil, true
	case "t":
		plan := application.PlanQuoteCenterSelection(application.QuoteCenterSelectionInput{
			Target: application.QuoteCenterTechnicals,
		})
		if plan.Allowed {
			m.quoteCenterMode = quoteCenterTechnicals
		}
		m.status = plan.Status
		if plan.LoadTechnical && m.needsTechnicalHistory(m.activeSymbol()) {
			return m, m.loadTechnicalHistoryCmd(m.activeSymbol()), true
		}
		return m, nil, true
	case "s":
		plan := application.PlanQuoteCenterSelection(application.QuoteCenterSelectionInput{
			Target:        application.QuoteCenterStatements,
			HasStatements: m.services.HasStatements(),
		})
		if !plan.Allowed {
			m.status = plan.Status
			return m, nil, true
		}
		m.quoteCenterMode = quoteCenterStatements
		m.status = plan.Status
		if plan.LoadStatement {
			return m, m.loadStatementCmd(m.activeSymbol()), true
		}
		return m, nil, true
	case "h":
		plan := application.PlanQuoteCenterSelection(application.QuoteCenterSelectionInput{
			Target:      application.QuoteCenterInsiders,
			HasInsiders: m.services.HasInsiders(),
		})
		if !plan.Allowed {
			m.status = plan.Status
			return m, nil, true
		}
		m.quoteCenterMode = quoteCenterInsiders
		m.status = plan.Status
		if plan.LoadInsiders {
			return m, m.loadInsidersCmd(m.activeSymbol()), true
		}
		return m, nil, true
	case "i":
		if m.quoteCenterMode == quoteCenterFilings {
			if m.aiRunning {
				return m, nil, true
			}
			item, ok := m.currentFiling()
			if !ok {
				return m, nil, true
			}
			prompt := filingAnalysisPrompt(m.activeSymbol(), item)
			m.tabIdx = tabAI
			m.helpOpen = false
			m.searchMode = false
			m.commandPaletteOpen = false
			m.aiPickerOpen = false
			m.aiFocused = false
			m.aiInput.Blur()
			m.pushAIUserMessage(prompt)
			m.aiRunning = true
			m.aiErr = nil
			m.status = "Loading selected filing for AI analysis…"
			return m, m.prepareFilingAnalysisCmd(m.activeSymbol(), item), true
		}
		if !m.quoteBottomPanelsVisible() {
			return m, nil, false
		}
		if !m.aiQuoteInsightRunning {
			m.aiQuoteInsightSymbol = strings.ToUpper(m.activeSymbol())
			m.aiQuoteInsightErr = nil
			m.aiQuoteInsightRunning = true
			m.status = "Refreshing quote AI insight…"
			return m, m.prepareQuoteInsightCmd(m.activeSymbol()), true
		}
		return m, nil, true
	case "d":
		if !m.quoteBottomPanelsVisible() {
			return m, nil, false
		}
		if len(m.config.Watchlist) > 1 {
			m.removeSelectedWatchlistSymbol()
			return m, tea.Batch(m.persistCmd(), m.loadAllCmd(m.activeSymbol())), true
		}
		return m, nil, true
	case "n":
		if !m.quoteBottomPanelsVisible() {
			return m, nil, false
		}
		m.cycleNewsSelection()
		return m, nil, true
	case "p":
		if !m.quoteBottomPanelsVisible() {
			return m, nil, false
		}
		m.cycleProfileScroll()
		return m, nil, true
	case "o":
		if !m.quoteBottomPanelsVisible() {
			return m, nil, false
		}
		if len(m.news) > 0 {
			_ = openURLFunc(m.news[m.newsSelected].URL)
			m.status = "Opened news item in browser"
		}
		return m, nil, true
	default:
		return m, nil, false
	}
}
