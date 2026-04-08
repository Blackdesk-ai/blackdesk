package tui

import (
	"fmt"
	"sort"
	"strings"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func marketNewsSourceRows(sources []domain.MarketNewsSource, items []domain.NewsItem) []marketTableRow {
	if len(sources) == 0 && len(items) == 0 {
		return []marketTableRow{{name: "Feeds", price: "-", chg: "-", styled: false}}
	}
	counts := make(map[string]int)
	for _, item := range items {
		name := strings.TrimSpace(item.Publisher)
		if name == "" {
			name = "Unknown"
		}
		counts[name]++
	}
	rows := make([]marketTableRow, 0, max(len(sources), len(counts)))
	seen := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		name := strings.TrimSpace(source.Name)
		if name == "" {
			continue
		}
		seen[name] = struct{}{}
		if counts[name] == 0 {
			continue
		}
		rows = append(rows, marketTableRow{name: name, price: fmt.Sprintf("%d", counts[name]), chg: ""})
	}
	extra := make([]string, 0, len(counts))
	for name := range counts {
		if _, ok := seen[name]; ok {
			continue
		}
		extra = append(extra, name)
	}
	sort.Strings(extra)
	for _, name := range extra {
		rows = append(rows, marketTableRow{name: name, price: fmt.Sprintf("%d", counts[name]), chg: ""})
	}
	if len(rows) == 0 {
		return []marketTableRow{{name: "Feeds", price: "-", chg: "-", styled: false}}
	}
	return rows
}

func compactSourceName(name string) string {
	switch strings.TrimSpace(name) {
	case "Federal Reserve":
		return "Fed"
	case "European Central Bank":
		return "ECB"
	case "Bank of England":
		return "BoE"
	case "Bureau of Economic Analysis":
		return "BEA"
	case "Financial Times":
		return "FT"
	case "Yahoo Finance":
		return "Yahoo"
	default:
		return name
	}
}

func marketBoardRows(m Model, items []marketBoardItem) []marketTableRow {
	rows := make([]marketTableRow, 0, len(items))
	for _, item := range items {
		quote, ok := m.lookupQuote(item.symbol)
		row := marketTableRow{name: item.label, price: "--", chg: "--"}
		if ok {
			displayPrice, displayChangePercent := marketDisplayQuoteLine(quote)
			row.price = ui.FormatMoney(displayPrice)
			row.chg = fmt.Sprintf("%+.2f%%", displayChangePercent)
			row.move = displayChangePercent
			row.styled = true
		}
		rows = append(rows, row)
	}
	return rows
}

func marketExtremeRows(m Model, best bool) []marketTableRow {
	boards := []struct {
		label string
		items []marketBoardItem
	}{
		{label: "US", items: marketUSBoard},
		{label: "Futures", items: marketFuturesBoard},
		{label: "Rates", items: marketRatesBoard},
		{label: "Macro", items: marketMacroBoard},
		{label: "Regions", items: marketRegionBoard},
	}
	rows := make([]marketTableRow, 0, len(boards))
	for _, board := range boards {
		item, quote, ok := extremeMarketMove(m, board.items, best)
		row := marketTableRow{name: board.label, price: "--", chg: "--"}
		if ok {
			displayPrice, displayChangePercent := marketDisplayQuoteLine(quote)
			row.name = item.label
			row.price = ui.FormatMoney(displayPrice)
			row.chg = fmt.Sprintf("%+.2f%%", displayChangePercent)
			row.move = displayChangePercent
			row.styled = true
		}
		rows = append(rows, row)
	}
	return rows
}

func (m Model) lookupQuote(symbol string) (domain.QuoteSnapshot, bool) {
	if strings.EqualFold(m.quote.Symbol, symbol) {
		return m.quote, true
	}
	quote, ok := m.watchQuotes[strings.ToUpper(symbol)]
	return quote, ok
}
