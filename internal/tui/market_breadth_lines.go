package tui

import (
	"fmt"
	"strings"
)

func marketBreadthLine(m Model) string {
	positive := 0
	negative := 0
	for _, item := range append(append(append([]marketBoardItem{}, marketUSBoard...), marketRegionBoard...), marketMacroBoard...) {
		quote, ok := m.lookupQuote(item.symbol)
		if !ok {
			continue
		}
		_, displayChangePercent := marketDisplayQuoteLine(quote)
		switch {
		case displayChangePercent > 0:
			positive++
		case displayChangePercent < 0:
			negative++
		}
	}
	if positive == 0 && negative == 0 {
		return "-"
	}
	return fmt.Sprintf("%d up / %d down", positive, negative)
}

func marketPressureLine(m Model) string {
	vix, vixOK := m.lookupQuote("^VIX")
	dollar, dollarOK := m.lookupQuote("DX-Y.NYB")
	tlt, tltOK := m.lookupQuote("TLT")
	if !vixOK && !dollarOK && !tltOK {
		return "-"
	}
	parts := make([]string, 0, 3)
	if vixOK {
		_, displayChangePercent := marketDisplayQuoteLine(vix)
		parts = append(parts, fmt.Sprintf("VIX %s", colorizeMarketChange(fmt.Sprintf("%+.2f%%", displayChangePercent), displayChangePercent, true)))
	}
	if dollarOK {
		_, displayChangePercent := marketDisplayQuoteLine(dollar)
		parts = append(parts, fmt.Sprintf("USD %s", colorizeMarketChange(fmt.Sprintf("%+.2f%%", displayChangePercent), displayChangePercent, true)))
	}
	if tltOK {
		_, displayChangePercent := marketDisplayQuoteLine(tlt)
		parts = append(parts, fmt.Sprintf("TLT %s", colorizeMarketChange(fmt.Sprintf("%+.2f%%", displayChangePercent), displayChangePercent, true)))
	}
	return strings.Join(parts, "  ")
}

func marketLeadershipRows(m Model) []string {
	boards := []struct {
		title string
		items []marketBoardItem
	}{
		{title: "US", items: marketUSBoard},
		{title: "Futures", items: marketFuturesBoard},
		{title: "Rates", items: marketRatesBoard},
		{title: "Macro", items: marketMacroBoard},
		{title: "Regions", items: marketRegionBoard},
	}
	rows := make([]string, 0, len(boards)*2)
	for _, board := range boards {
		best := bestMarketMoveLine(m, board.items)
		worst := worstMarketMoveLine(m, board.items)
		rows = append(rows, fmt.Sprintf("%-8s best  %s", board.title, best))
		rows = append(rows, fmt.Sprintf("%-8s worst %s", board.title, worst))
	}
	return rows
}
