package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
)

func (m Model) statementBundleForSymbol(symbol string) aiStatementsSnapshot {
	return aiStatementsSnapshot{
		IncomeAnnual:          m.statementForBundle(symbol, domain.StatementKindIncome, domain.StatementFrequencyAnnual),
		BalanceSheetAnnual:    m.statementForBundle(symbol, domain.StatementKindBalanceSheet, domain.StatementFrequencyAnnual),
		CashFlowAnnual:        m.statementForBundle(symbol, domain.StatementKindCashFlow, domain.StatementFrequencyAnnual),
		IncomeQuarterly:       m.statementForBundle(symbol, domain.StatementKindIncome, domain.StatementFrequencyQuarterly),
		BalanceSheetQuarterly: m.statementForBundle(symbol, domain.StatementKindBalanceSheet, domain.StatementFrequencyQuarterly),
		CashFlowQuarterly:     m.statementForBundle(symbol, domain.StatementKindCashFlow, domain.StatementFrequencyQuarterly),
	}
}

func (m Model) insidersForSymbol(symbol string) domain.InsiderSnapshot {
	data, ok := m.cachedInsiders(symbol)
	if !ok {
		return domain.InsiderSnapshot{}
	}
	return data
}

func (m Model) statementForBundle(symbol string, kind domain.StatementKind, freq domain.StatementFrequency) domain.FinancialStatement {
	data, ok := m.cachedStatement(symbol, kind, freq)
	if !ok {
		return domain.FinancialStatement{}
	}
	return data
}

func (m Model) aiQuoteStats() map[string]string {
	stats := make(map[string]string, 9)
	for _, line := range overviewStatsGrid(lipgloss.NewStyle(), 96, m.quote, m.fundamentals) {
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) < 2 {
			continue
		}
		stats[strings.Join(parts[:len(parts)-1], " ")] = parts[len(parts)-1]
	}
	if earningsYield, ok := earningsYieldValue(m.quote, m.fundamentals); ok {
		stats["Earnings Yield"] = formatOptionalPercent(earningsYield, true)
	}
	if growthEst := impliedEPSGrowthEstimateText(m.fundamentals); growthEst != "-" {
		stats["Fwd Growth"] = growthEst
	}
	return stats
}

func (m Model) aiTechnicalSections() map[string][]aiStatRow {
	series := m.technicalSeries(m.activeSymbol())
	if len(series.Candles) == 0 {
		series = m.series
	}
	snapshot := buildTechnicalSnapshot(m.quote, series)
	return map[string][]aiStatRow{
		"momentum":   aiRowsFromMarketRows(technicalMomentumRows(snapshot)),
		"trend":      aiRowsFromMarketRows(technicalTrendRows(snapshot)),
		"stat_edge":  aiRowsFromMarketRows(technicalStatRows(snapshot)),
		"volatility": aiRowsFromMarketRows(technicalVolatilityRows(snapshot)),
		"volume":     aiRowsFromMarketRows(technicalVolumeRows(snapshot)),
	}
}

func (m Model) aiMarketSections() map[string][]aiStatRow {
	return map[string][]aiStatRow{
		"futures":     aiRowsFromMarketRows(marketBoardRows(m, marketFuturesBoard)),
		"yields":      aiRowsFromMarketRows(marketBoardRows(m, marketYieldBoard)),
		"credit":      aiRowsFromMarketRows(marketBoardRows(m, marketRatesBoard)),
		"commodities": aiRowsFromMarketRows(marketBoardRows(m, marketMacroBoard)),
		"fx":          aiRowsFromMarketRows(marketBoardRows(m, marketFXBoard)),
		"volatility":  aiRowsFromMarketRows(marketBoardRows(m, marketVolBoard)),
		"regions":     aiRowsFromMarketRows(marketBoardRows(m, marketRegionBoard)),
		"countries":   aiRowsFromMarketRows(marketBoardRows(m, marketCountryBoard)),
		"sectors":     aiRowsFromMarketRows(marketBoardRows(m, marketSectorBoard)),
	}
}

func aiMarketRegimeFromSnapshot(risk domain.MarketRiskSnapshot) aiMarketRegime {
	if !risk.Available || risk.Min >= risk.Max {
		return aiMarketRegime{}
	}
	out := aiMarketRegime{
		Available:      true,
		Score:          risk.Score,
		Min:            risk.Min,
		Max:            risk.Max,
		Stance:         marketRiskStanceKey(risk),
		Strength:       marketRiskStrength(risk),
		Interpretation: "score 0 is neutral, scores above 0 are risk-on, scores below 0 are risk-off, and larger absolute values mean stronger conviction",
		SourceLabel:    strings.TrimSpace(risk.Label),
		Thresholds: aiMarketRiskThresholds{
			SMABufferPct:    risk.Thresholds.SMABufferPct,
			Breadth50Buffer: risk.Thresholds.Breadth50Buffer,
		},
		Components:     cloneAIMarketRiskComponents(risk.Components),
		Inputs:         cloneAIMarketRiskInputs(risk.Inputs),
		MarketTimezone: strings.TrimSpace(risk.MarketZone),
		MarketCalendar: strings.TrimSpace(risk.MarketCalendar),
	}
	if !risk.GeneratedAt.IsZero() {
		out.GeneratedAt = risk.GeneratedAt.Format(time.RFC3339)
	}
	if !risk.MarketNow.IsZero() {
		out.MarketNow = risk.MarketNow.Format(time.RFC3339)
	}
	return out
}

func cloneAIMarketRiskComponents(src map[string]int) map[string]int {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]int, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func cloneAIMarketRiskInputs(src map[string]domain.MarketRiskInput) map[string]aiMarketRiskInput {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]aiMarketRiskInput, len(src))
	for key, value := range src {
		out[key] = aiMarketRiskInput{
			Name:    value.Name,
			Symbol:  value.Symbol,
			Current: value.Current,
			SMA200:  value.SMA200,
		}
	}
	return out
}

func (m Model) aiTechnicalLookup() map[string]aiStatRow {
	lookup := make(map[string]aiStatRow)
	for _, rows := range m.aiTechnicalSections() {
		for _, row := range rows {
			key := normalizeIndicatorKey(row.Name)
			if key == "" {
				continue
			}
			lookup[key] = row
		}
	}
	return lookup
}

func (m Model) aiTechnicalValues() map[string]string {
	values := make(map[string]string)
	for key, row := range m.aiTechnicalLookup() {
		values[key] = row.Value
	}
	return values
}

func normalizeIndicatorKey(input string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "", "/", "", "%", "", ".", "")
	return strings.ToUpper(replacer.Replace(strings.TrimSpace(input)))
}

func (m Model) statusMetaMarketSource() string {
	if strings.TrimSpace(m.services.ActiveProviderName()) == "" {
		return "unknown"
	}
	return m.services.ActiveProviderName()
}
