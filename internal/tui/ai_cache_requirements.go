package tui

import (
	"strings"

	"blackdesk/internal/application"
	"blackdesk/internal/domain"
)

func aiFundamentalsMissing(f domain.FundamentalsSnapshot) bool {
	return strings.TrimSpace(f.Symbol) == "" &&
		strings.TrimSpace(f.Description) == "" &&
		strings.TrimSpace(f.Sector) == "" &&
		f.MarketCap == 0 &&
		f.Revenue == 0 &&
		f.TotalCash == 0
}

func (m Model) missingAIQuoteSymbols(activeSymbol string) []string {
	candidates := append(
		application.WatchlistQuoteSymbols(m.config.Watchlist, activeSymbol),
		application.SupplementalMarketQuoteSymbols(m.config.Watchlist, activeSymbol, marketDashboardSymbols)...,
	)
	symbols := make([]string, 0, len(candidates))
	for _, symbol := range candidates {
		if _, ok := m.watchQuotes[strings.ToUpper(symbol)]; ok {
			continue
		}
		symbols = append(symbols, symbol)
	}
	return symbols
}
