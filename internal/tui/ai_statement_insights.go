package tui

func aiStatementInsights(bundle aiStatementsSnapshot) []aiStatRow {
	rows := make([]aiStatRow, 0, 8)
	if row, ok := statementGrowthInsight("Revenue YoY", bundle.IncomeAnnual, "TotalRevenue"); ok {
		rows = append(rows, row)
	}
	if row, ok := statementGrowthInsight("Net Income YoY", bundle.IncomeAnnual, "NetIncome"); ok {
		rows = append(rows, row)
	}
	if row, ok := statementGrowthInsight("Revenue YoY (Q)", bundle.IncomeQuarterly, "TotalRevenue"); ok {
		rows = append(rows, row)
	}
	if row, ok := statementGrowthInsight("Net Income YoY (Q)", bundle.IncomeQuarterly, "NetIncome"); ok {
		rows = append(rows, row)
	}
	if row, ok := statementGrowthInsight("OCF YoY", bundle.CashFlowAnnual, "OperatingCashFlow"); ok {
		rows = append(rows, row)
	}
	if row, ok := statementGrowthInsight("FCF YoY", bundle.CashFlowAnnual, "FreeCashFlow"); ok {
		rows = append(rows, row)
	}
	if row, ok := statementBalanceInsight(bundle.BalanceSheetAnnual); ok {
		rows = append(rows, row)
	}
	if row, ok := statementLiquidityInsight(bundle.BalanceSheetAnnual); ok {
		rows = append(rows, row)
	}
	return rows
}
