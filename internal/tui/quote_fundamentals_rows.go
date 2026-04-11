package tui

import (
	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func quoteFundamentalsValuationRows(quote domain.QuoteSnapshot, f domain.FundamentalsSnapshot) []marketTableRow {
	marketCap := f.MarketCap
	if marketCap == 0 {
		marketCap = quote.MarketCap
	}
	return []marketTableRow{
		{name: "Market cap", price: ui.FormatCompactInt(marketCap), chg: "", move: 0, styled: false},
		{name: "Ent value", price: formatCompactIntDash(f.EnterpriseValue), chg: "", move: 0, styled: false},
		{name: "Trailing PE", price: formatMetricFloat(f.TrailingPE), chg: "", move: 0, styled: false},
		{name: "Forward PE", price: formatMetricFloat(f.ForwardPE), chg: "", move: 0, styled: false},
		{name: "PEG", price: formatMetricFloat(pegRatioValue(quote, f)), chg: "", move: 0, styled: false},
		{name: "P/B", price: formatMetricFloat(f.PriceToBook), chg: "", move: 0, styled: false},
		{name: "P/S", price: formatMetricFloat(f.PriceToSales), chg: "", move: 0, styled: false},
		{name: "EV/EBITDA", price: formatMetricFloat(f.EnterpriseToEBITDA), chg: "", move: 0, styled: false},
		{name: "Dividend", price: percentDash(f.DividendYield), chg: yieldBadge(f.DividendYield), move: f.DividendYield, styled: f.DividendYield > 0},
	}
}

func quoteFundamentalsProfitabilityRows(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) []marketTableRow {
	return []marketTableRow{
		{name: "Gross margin", price: percentDash(f.GrossMargins), chg: "", move: 0, styled: false},
		{name: "Operating", price: percentDash(f.OperatingMargins), chg: "", move: 0, styled: false},
		{name: "Profit margin", price: percentDash(f.ProfitMargins), chg: "", move: 0, styled: false},
		{name: "ROIC", price: percentDash(f.ReturnOnInvestedCapital), chg: "", move: 0, styled: false},
		{name: "ROE", price: percentDash(f.ReturnOnEquity), chg: "", move: 0, styled: false},
		{name: "ROA", price: percentDash(f.ReturnOnAssets), chg: "", move: 0, styled: false},
		{name: "Rev growth", price: percentDash(f.RevenueGrowth), chg: "", move: 0, styled: false},
		{name: "EPS growth", price: percentDash(f.EarningsGrowth), chg: "", move: 0, styled: false},
		{name: "Growth Est.", price: impliedEPSGrowthBandText(q, f), chg: "", move: 0, styled: false},
	}
}

func quoteFundamentalsFinancialRows(f domain.FundamentalsSnapshot) []marketTableRow {
	netCash := f.TotalCash - f.TotalDebt
	return []marketTableRow{
		{name: "Revenue", price: formatCompactIntDash(f.Revenue), chg: "", move: 0, styled: false},
		{name: "EBITDA", price: formatCompactIntDash(f.EBITDA), chg: "", move: 0, styled: false},
		{name: "Op cash flow", price: formatCompactIntDash(f.OperatingCashflow), chg: "", move: 0, styled: false},
		{name: "Free cash flow", price: formatCompactIntDash(f.FreeCashflow), chg: cashFlowSignal(f.FreeCashflow), move: float64(f.FreeCashflow), styled: f.FreeCashflow != 0},
		{name: "Cash", price: formatCompactIntDash(f.TotalCash), chg: "", move: 0, styled: false},
		{name: "Debt", price: formatCompactIntDash(f.TotalDebt), chg: "", move: 0, styled: false},
		{name: "Net cash", price: formatCompactIntDash(netCash), chg: netCashSignal(netCash), move: float64(netCash), styled: netCash != 0},
		{name: "Debt/equity", price: formatMetricFloat(f.DebtToEquity), chg: "", move: 0, styled: false},
		{name: "Current", price: formatMetricFloat(f.CurrentRatio), chg: "", move: 0, styled: false},
		{name: "Rev/share", price: formatMoneyDash(f.RevenuePerShare), chg: "", move: 0, styled: false},
	}
}

func splitFinancialFundamentalsRows(rows []marketTableRow) ([]marketTableRow, []marketTableRow) {
	if len(rows) <= 4 {
		return rows, nil
	}
	return rows[:4], rows[4:]
}

func splitFundamentalsRows(rows []marketTableRow) ([]marketTableRow, []marketTableRow) {
	if len(rows) <= 1 {
		return rows, nil
	}
	mid := (len(rows) + 1) / 2
	return rows[:mid], rows[mid:]
}

func quoteFundamentalsAnalystRows(quote domain.QuoteSnapshot, f domain.FundamentalsSnapshot) []marketTableRow {
	upside := analystUpsideValue(f, quote.Price)
	return []marketTableRow{
		{name: "Rating", price: recommendationBadge(f.RecommendationKey), chg: "", move: 0, styled: false},
		{name: "Mean score", price: formatAnalystMean(f.RecommendationMean), chg: "", move: 0, styled: false},
		{name: "Opinions", price: formatAnalystOpinions(f.AnalystOpinions), chg: "", move: 0, styled: false},
		{name: "Target low", price: formatMoneyDash(f.TargetLowPrice), chg: "", move: 0, styled: false},
		{name: "Target mean", price: formatMoneyDash(f.TargetMeanPrice), chg: analystUpsideLine(f, quote.Price), move: upside, styled: upside != 0},
		{name: "Target high", price: formatMoneyDash(f.TargetHighPrice), chg: "", move: 0, styled: false},
	}
}
