package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/application"
	"blackdesk/internal/domain"
)

func (m Model) loadWatchlistQuotesCmd(activeSymbol string) tea.Cmd {
	symbols := application.WatchlistQuoteSymbols(m.config.Watchlist, activeSymbol)
	if len(symbols) == 0 {
		return nil
	}
	return func() tea.Msg {
		quotes, err := m.services.GetQuotes(m.ctx, symbols)
		return quotesLoadedMsg{quotes: quotes, err: err}
	}
}

func (m Model) loadMarketQuotesCmd() tea.Cmd {
	symbols := application.SupplementalMarketQuoteSymbols(m.config.Watchlist, m.activeSymbol(), marketDashboardSymbols)
	if len(symbols) == 0 {
		return nil
	}
	return func() tea.Msg {
		quotes, err := m.services.GetQuotes(m.ctx, symbols)
		return quotesLoadedMsg{quotes: quotes, err: err}
	}
}

func (m Model) loadMarketNewsCmd() tea.Cmd {
	if m.services == nil {
		return nil
	}
	return func() tea.Msg {
		items, srcs, err := m.services.GetMarketNews(m.ctx)
		return marketNewsLoadedMsg{items: items, srcs: srcs, err: err}
	}
}

func (m Model) loadScreenerCmd(userTriggered bool) tea.Cmd {
	if !m.services.HasScreeners() || len(m.screenerDefs) == 0 {
		return nil
	}
	def := m.currentScreenerDefinition()
	if strings.TrimSpace(def.ID) == "" {
		return nil
	}
	return func() tea.Msg {
		data, err := m.services.GetScreener(m.ctx, def.ID, screenerResultCount)
		return screenerLoadedMsg{data: data, err: err, userTriggered: userTriggered}
	}
}

func filterMarketNewsRecent(items []domain.NewsItem, now time.Time) []domain.NewsItem {
	return application.FilterRecentMarketNews(items, now)
}

func marketNewsIdentity(item domain.NewsItem) string {
	return application.MarketNewsIdentity(item)
}
