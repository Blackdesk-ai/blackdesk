package tui

import "blackdesk/internal/domain"

func marketBasketScore(m Model, items []marketBoardItem) float64 {
	if len(items) == 0 {
		return 0
	}
	total := 0.0
	count := 0
	for _, item := range items {
		quote, ok := m.lookupQuote(item.symbol)
		if !ok {
			continue
		}
		_, displayChangePercent := marketDisplayQuoteLine(quote)
		total += clampFloat((displayChangePercent+2)/4, 0, 1)
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func marketDisplayQuoteLine(quote domain.QuoteSnapshot) (price, changePercent float64) {
	displayPrice, _, displayChangePercent, _ := displayQuoteLine(quote)
	return displayPrice, displayChangePercent
}
