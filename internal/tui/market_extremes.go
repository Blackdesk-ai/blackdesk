package tui

import (
	"fmt"
	"math"

	"blackdesk/internal/domain"
)

func bestMarketMoveLine(m Model, items []marketBoardItem) string {
	item, quote, ok := extremeMarketMove(m, items, true)
	if !ok {
		return "-"
	}
	_, displayChangePercent := marketDisplayQuoteLine(quote)
	return fmt.Sprintf("%s %s", item.label, colorizeMarketChange(fmt.Sprintf("%+.2f%%", displayChangePercent), displayChangePercent, true))
}

func worstMarketMoveLine(m Model, items []marketBoardItem) string {
	item, quote, ok := extremeMarketMove(m, items, false)
	if !ok {
		return "-"
	}
	_, displayChangePercent := marketDisplayQuoteLine(quote)
	return fmt.Sprintf("%s %s", item.label, colorizeMarketChange(fmt.Sprintf("%+.2f%%", displayChangePercent), displayChangePercent, true))
}

func extremeMarketMove(m Model, items []marketBoardItem, best bool) (marketBoardItem, domain.QuoteSnapshot, bool) {
	var chosen marketBoardItem
	var chosenQuote domain.QuoteSnapshot
	chosenChangePercent := 0.0
	found := false
	for _, item := range items {
		quote, ok := m.lookupQuote(item.symbol)
		if !ok {
			continue
		}
		_, displayChangePercent := marketDisplayQuoteLine(quote)
		if !found || (best && displayChangePercent > chosenChangePercent) || (!best && displayChangePercent < chosenChangePercent) {
			chosen = item
			chosenQuote = quote
			chosenChangePercent = displayChangePercent
			found = true
		}
	}
	return chosen, chosenQuote, found
}

func focusAssetLine(m Model) string {
	item, quote, ok := extremeMarketMove(m, append(append([]marketBoardItem{}, marketMacroBoard...), marketRegionBoard...), true)
	if !ok {
		return "-"
	}
	_, displayChangePercent := marketDisplayQuoteLine(quote)
	if math.Abs(displayChangePercent) < 1 {
		item, quote, ok = extremeMarketMove(m, append(append([]marketBoardItem{}, marketMacroBoard...), marketRegionBoard...), false)
		if !ok {
			return "-"
		}
	}
	_, displayChangePercent = marketDisplayQuoteLine(quote)
	return fmt.Sprintf("%s %s", item.label, colorizeMarketChange(fmt.Sprintf("%+.2f%%", displayChangePercent), displayChangePercent, true))
}
