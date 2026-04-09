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
	data domain.FundamentalsSnapshot
	err  error
}

type statementLoadedMsg struct {
	data domain.FinancialStatement
	err  error
}

type insidersLoadedMsg struct {
	data domain.InsiderSnapshot
	err  error
}

type searchLoadedMsg struct {
	results []domain.SymbolRef
	err     error
}
