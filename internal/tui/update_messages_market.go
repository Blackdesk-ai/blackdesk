package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
	"blackdesk/internal/domain"
)

func (m Model) handleQuoteLoaded(msg quoteLoadedMsg) (Model, tea.Cmd) {
	if msg.err == nil && msg.symbol != "" {
		m.watchQuotes[strings.ToUpper(msg.symbol)] = msg.quote
	}
	if strings.EqualFold(msg.symbol, m.activeSymbol()) {
		m.quote = msg.quote
		m.errQuote = msg.err
		m.status = fmt.Sprintf("Loaded quote for %s", m.quote.Symbol)
		m.lastUpdated = time.Now()
	}
	return m, nil
}

func (m Model) handleQuotesLoaded(msg quotesLoadedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.status = msg.err.Error()
		return m, nil
	}
	for _, quote := range msg.quotes {
		if quote.Symbol == "" {
			continue
		}
		m.watchQuotes[strings.ToUpper(quote.Symbol)] = quote
	}
	if m.pendingMarketOpinionRefresh && !m.aiMarketOpinionRunning && m.hasSufficientMarketOpinionData() {
		m.pendingMarketOpinionRefresh = false
		m.aiMarketOpinionRunning = true
		return m, m.runMarketOpinionCmd()
	}
	return m, nil
}

func (m Model) handleHistoryLoaded(msg historyLoadedMsg) (Model, tea.Cmd) {
	if !strings.EqualFold(msg.symbol, m.activeSymbol()) {
		return m, nil
	}
	m.series = msg.series
	m.errHistory = msg.err
	return m, nil
}

func (m Model) handleTechnicalHistoryLoaded(msg technicalHistoryLoadedMsg) (Model, tea.Cmd) {
	if msg.err == nil && msg.series.Symbol != "" {
		m.technicalCache[strings.ToUpper(msg.series.Symbol)] = msg.series
	}
	if !strings.EqualFold(msg.symbol, m.activeSymbol()) {
		return m, nil
	}
	m.errTechnicalHistory = msg.err
	return m, nil
}

func (m Model) handleNewsLoaded(msg newsLoadedMsg) (Model, tea.Cmd) {
	m.news = msg.items
	m.errNews = msg.err
	if m.newsSelected >= len(m.news) {
		m.newsSelected = 0
	}
	return m, nil
}

func (m Model) handleMarketNewsLoaded(msg marketNewsLoadedMsg) (Model, tea.Cmd) {
	now := time.Now()
	result := application.MergeMarketNews(application.MarketNewsMergeInput{
		PreviousItems:   m.marketNews,
		PreviousSeen:    m.marketNewsSeen,
		PreviousSources: m.marketNewsSources,
		IncomingItems:   msg.items,
		IncomingSources: msg.srcs,
		SelectedIndex:   m.marketNewsSel,
		Now:             now,
		Err:             msg.err,
	})
	m.marketNews = result.Items
	m.marketNewsSeen = result.Seen
	m.marketNewsFresh = result.Fresh
	m.marketNewsSources = result.Sources
	m.errMarketNews = msg.err
	if !result.UpdatedAt.IsZero() {
		m.marketNewsUpdated = result.UpdatedAt
	}
	if !result.LastRefresh.IsZero() {
		m.lastMarketNews = result.LastRefresh
	}
	if result.ApplyStatus {
		m.status = result.Status
	}
	m.marketNewsSel = result.SelectedIndex
	m.ensureMarketNewsSelectionVisible()
	return m, nil
}

func (m Model) handleMarketRiskLoaded(msg marketRiskLoadedMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.marketRisk = domain.MarketRiskSnapshot{}
		return m, nil
	}
	m.marketRisk = msg.data
	return m, nil
}

func (m Model) handleTick(msg tickMsg) (Model, tea.Cmd) {
	now := time.Time(msg)
	m.clock = now
	plan := application.PlanAutoRefresh(application.AutoRefreshInput{
		Now:                   now,
		LastAutoRefresh:       m.lastAutoRefresh,
		LastMarketNewsRefresh: m.lastMarketNews,
		RefreshSeconds:        m.config.RefreshSeconds,
		DefaultRefreshSeconds: domain.DefaultRefreshSeconds,
		NewsTabActive:         m.tabIdx == tabNews,
		ScreenerTabActive:     m.tabIdx == tabScreener,
		ScreenerLoaded:        m.screenerLoaded,
		MarketNewsInterval:    marketNewsRefreshInterval,
	})
	m.lastAutoRefresh = plan.NextLastAutoRefresh
	m.lastMarketNews = plan.NextLastMarketRefresh
	cmds := []tea.Cmd{tickCmd(time.Second)}
	if plan.RefreshAll {
		cmds = append(cmds, m.autoRefreshQuotesCmd(m.activeSymbol()))
	}
	return m, tea.Batch(cmds...)
}
