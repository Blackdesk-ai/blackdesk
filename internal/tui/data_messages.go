package tui

import "blackdesk/internal/domain"

type quoteLoadedMsg struct {
	symbol string
	quote  domain.QuoteSnapshot
	err    error
}

type quotesLoadedMsg struct {
	quotes []domain.QuoteSnapshot
	err    error
}

type historyLoadedMsg struct {
	series domain.PriceSeries
	err    error
}

type technicalHistoryLoadedMsg struct {
	series domain.PriceSeries
	err    error
}

type newsLoadedMsg struct {
	items []domain.NewsItem
	err   error
}

type marketNewsLoadedMsg struct {
	items []domain.NewsItem
	srcs  []domain.MarketNewsSource
	err   error
}

type marketRiskLoadedMsg struct {
	data domain.MarketRiskSnapshot
	err  error
}

type screenerLoadedMsg struct {
	data          domain.ScreenerResult
	err           error
	userTriggered bool
}

type fundamentalsLoadedMsg struct {
	symbol string
	data   domain.FundamentalsSnapshot
	err    error
}

type statementLoadedMsg struct {
	data domain.FinancialStatement
	err  error
}

type insidersLoadedMsg struct {
	data domain.InsiderSnapshot
	err  error
}

type ownersLoadedMsg struct {
	data domain.OwnershipSnapshot
	err  error
}

type analystRecommendationsLoadedMsg struct {
	data domain.AnalystRecommendationsSnapshot
	err  error
}

type filingsLoadedMsg struct {
	data domain.FilingsSnapshot
	err  error
}

type earningsLoadedMsg struct {
	data domain.EarningsSnapshot
	err  error
}

type calendarLoadedMsg struct {
	filter calendarFilterMode
	data   domain.EconomicCalendarSnapshot
	err    error
}

type searchDebouncedMsg struct {
	id    int
	query string
}

type searchLoadedMsg struct {
	id      int
	query   string
	results []domain.SymbolRef
	err     error
}

type commandPaletteDebouncedMsg struct {
	id    int
	query string
}

type commandPaletteLoadedMsg struct {
	id      int
	query   string
	results []domain.SymbolRef
	err     error
}
