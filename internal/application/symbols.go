package application

import "strings"

func WatchlistQuoteSymbols(watchlist []string, activeSymbol string) []string {
	out := make([]string, 0, len(watchlist))
	for _, symbol := range watchlist {
		if strings.EqualFold(symbol, activeSymbol) {
			continue
		}
		out = append(out, symbol)
	}
	return out
}

func SupplementalMarketQuoteSymbols(watchlist []string, activeSymbol string, marketSymbols []string) []string {
	out := make([]string, 0, len(marketSymbols))
	seen := make(map[string]struct{}, len(marketSymbols)+len(watchlist)+1)
	for _, symbol := range watchlist {
		seen[strings.ToUpper(symbol)] = struct{}{}
	}
	if trimmed := strings.ToUpper(strings.TrimSpace(activeSymbol)); trimmed != "" {
		seen[trimmed] = struct{}{}
	}
	for _, symbol := range marketSymbols {
		key := strings.ToUpper(strings.TrimSpace(symbol))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, symbol)
	}
	return out
}
