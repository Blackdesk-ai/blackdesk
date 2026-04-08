package tui

import (
	"fmt"
	"math"

	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
)

func statementGrowthInsight(name string, stmt domain.FinancialStatement, key string) (aiStatRow, bool) {
	values, ok := statementRowValues(stmt, key)
	if !ok {
		return aiStatRow{}, false
	}
	changePct, ok := statementGrowthPercent(stmt, values, 0)
	if !ok {
		return aiStatRow{}, false
	}
	signal := "flat"
	switch {
	case changePct > 0:
		signal = "improving"
	case changePct < 0:
		signal = "deteriorating"
	}
	return aiStatRow{
		Name:   name,
		Value:  fmt.Sprintf("%+.1f%%", changePct),
		Signal: signal,
	}, true
}

func statementBalanceInsight(stmt domain.FinancialStatement) (aiStatRow, bool) {
	if value, ok := statementLatestValue(stmt, "NetDebt"); ok {
		signal := "net cash"
		if value > 0 {
			signal = "net debt"
		}
		return aiStatRow{
			Name:   "Net Debt",
			Value:  formatStatementInsightAmount(value),
			Signal: signal,
		}, true
	}

	debt, debtOK := statementLatestValue(stmt, "TotalDebt")
	cash, cashOK := statementLatestValueAny(stmt, []string{"CashCashEquivalentsAndShortTermInvestments", "CashAndCashEquivalents"})
	if !debtOK || !cashOK {
		return aiStatRow{}, false
	}
	netDebt := debt - cash
	signal := "net cash"
	if netDebt > 0 {
		signal = "net debt"
	}
	return aiStatRow{
		Name:   "Net Debt",
		Value:  formatStatementInsightAmount(netDebt),
		Signal: signal,
	}, true
}

func statementLiquidityInsight(stmt domain.FinancialStatement) (aiStatRow, bool) {
	totalDebt, debtOK := statementLatestValue(stmt, "TotalDebt")
	totalAssets, assetsOK := statementLatestValue(stmt, "TotalAssets")
	if !debtOK || !assetsOK || totalAssets == 0 {
		return aiStatRow{}, false
	}
	ratioPct := (totalDebt / totalAssets) * 100
	signal := "moderate leverage"
	switch {
	case ratioPct < 20:
		signal = "light leverage"
	case ratioPct > 50:
		signal = "heavy leverage"
	}
	return aiStatRow{
		Name:   "Debt / Assets",
		Value:  ui.FormatPercent(ratioPct),
		Signal: signal,
	}, true
}

func statementRowValues(stmt domain.FinancialStatement, key string) ([]domain.StatementValue, bool) {
	for _, row := range stmt.Rows {
		if row.Key == key {
			return row.Values, true
		}
	}
	return nil, false
}

func statementLatestValue(stmt domain.FinancialStatement, key string) (float64, bool) {
	values, ok := statementRowValues(stmt, key)
	if !ok || len(values) == 0 || !values[0].Present {
		return 0, false
	}
	return values[0].Value, true
}

func statementLatestValueAny(stmt domain.FinancialStatement, keys []string) (float64, bool) {
	for _, key := range keys {
		if value, ok := statementLatestValue(stmt, key); ok {
			return value, true
		}
	}
	return 0, false
}

func formatStatementInsightAmount(v float64) string {
	if v == 0 {
		return "0"
	}
	return ui.FormatCompactInt(int64(math.Round(v)))
}
