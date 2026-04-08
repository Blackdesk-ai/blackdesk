package tui

import "fmt"

func volatilityCurveLine(m Model) string {
	vix9d, ok9d := m.lookupQuote("^VIX9D")
	vix, okVIX := m.lookupQuote("^VIX")
	vix3m, ok3m := m.lookupQuote("^VIX3M")
	if !ok9d || !okVIX || !ok3m {
		return "-"
	}
	vix9dPrice, _ := marketDisplayQuoteLine(vix9d)
	vixPrice, _ := marketDisplayQuoteLine(vix)
	vix3mPrice, _ := marketDisplayQuoteLine(vix3m)
	switch {
	case vix9dPrice > vixPrice && vixPrice >= vix3mPrice:
		return marketMoveStyle(-1).Render("inverted")
	case vix9dPrice < vixPrice && vixPrice < vix3mPrice:
		return marketMoveStyle(1).Render("normal")
	default:
		return marketMoveStyle(0).Render("flat")
	}
}

func curve2s10sLine(m Model) string {
	twoYear, ok2 := m.lookupQuote("2YY=F")
	tenYear, ok10 := m.lookupQuote("^TNX")
	if !ok2 || !ok10 {
		return "-"
	}
	twoYearPrice, _ := marketDisplayQuoteLine(twoYear)
	tenYearPrice, _ := marketDisplayQuoteLine(tenYear)
	spread := tenYearPrice - twoYearPrice
	return colorizeMarketChange(fmt.Sprintf("%+.2fbp", spread*100), spread, true)
}
