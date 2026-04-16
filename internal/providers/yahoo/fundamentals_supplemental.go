package yahoo

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

var supplementalIncomeFields = []string{
	"TotalRevenue",
	"GrossProfit",
	"EBIT",
	"OperatingIncome",
	"PretaxIncome",
	"TaxProvision",
	"NetIncome",
}

var supplementalBalanceFields = []string{
	"InvestedCapital",
	"TotalAssets",
	"StockholdersEquity",
	"TotalDebt",
	"CashCashEquivalentsAndShortTermInvestments",
	"CashAndCashEquivalents",
}

var supplementalCashFlowFields = []string{
	"OperatingCashFlow",
	"CashFlowFromContinuingOperatingActivities",
	"CapitalExpenditure",
	"FreeCashFlow",
}

type datedFundamentalsValue struct {
	date  string
	end   time.Time
	value float64
}

type supplementalProfitability struct {
	grossMargin     float64
	operatingMargin float64
	profitMargin    float64
	returnOnAssets  float64
	returnOnEquity  float64
	returnOnIC      float64
	investedCapital float64
}

type supplementalCashFlow struct {
	operatingCashFlow float64
	freeCashFlow      float64
}

func (p *Provider) hydrateSupplementalFundamentals(ctx context.Context, symbol string, fundamentals *domain.FundamentalsSnapshot) {
	if fundamentals == nil {
		return
	}

	incomeQuarterlyResp, err := p.fetchSupplementalTimeseries(ctx, symbol, "income-quarterly", domain.StatementFrequencyQuarterly, supplementalIncomeFields)
	if err != nil {
		return
	}
	incomeAnnualResp, err := p.fetchSupplementalTimeseries(ctx, symbol, "income-annual", domain.StatementFrequencyAnnual, supplementalIncomeFields)
	if err != nil {
		return
	}
	balanceQuarterlyResp, err := p.fetchSupplementalTimeseries(ctx, symbol, "balance-quarterly", domain.StatementFrequencyQuarterly, supplementalBalanceFields)
	if err != nil {
		return
	}
	balanceAnnualResp, err := p.fetchSupplementalTimeseries(ctx, symbol, "balance-annual", domain.StatementFrequencyAnnual, supplementalBalanceFields)
	if err != nil {
		return
	}
	cashFlowQuarterlyResp, err := p.fetchSupplementalTimeseries(ctx, symbol, "cashflow-quarterly", domain.StatementFrequencyQuarterly, supplementalCashFlowFields)
	if err != nil {
		return
	}
	cashFlowAnnualResp, err := p.fetchSupplementalTimeseries(ctx, symbol, "cashflow-annual", domain.StatementFrequencyAnnual, supplementalCashFlowFields)
	if err != nil {
		return
	}
	resp := mergeStatementTimeseriesResponses(
		incomeQuarterlyResp,
		incomeAnnualResp,
		balanceQuarterlyResp,
		balanceAnnualResp,
		cashFlowQuarterlyResp,
		cashFlowAnnualResp,
	)

	investedCapital, roic, ok := deriveSupplementalROIC(resp)
	if ok {
		fundamentals.InvestedCapital = int64(math.Round(investedCapital))
		fundamentals.ReturnOnInvestedCapital = roic
	}

	if metrics, ok := deriveSupplementalProfitability(resp); ok {
		fundamentals.GrossMargins = metrics.grossMargin
		fundamentals.OperatingMargins = metrics.operatingMargin
		fundamentals.ProfitMargins = metrics.profitMargin
		fundamentals.ReturnOnAssets = metrics.returnOnAssets
		fundamentals.ReturnOnEquity = metrics.returnOnEquity
		fundamentals.ReturnOnInvestedCapital = metrics.returnOnIC
		if metrics.investedCapital > 0 {
			fundamentals.InvestedCapital = int64(math.Round(metrics.investedCapital))
		}
	}

	if metrics, ok := deriveSupplementalCashFlow(resp); ok {
		if metrics.operatingCashFlow != 0 {
			fundamentals.OperatingCashflow = int64(math.Round(metrics.operatingCashFlow))
		}
		if metrics.freeCashFlow != 0 {
			fundamentals.FreeCashflow = int64(math.Round(metrics.freeCashFlow))
		}
	}
}

func (p *Provider) fetchSupplementalTimeseries(ctx context.Context, symbol, key string, frequency domain.StatementFrequency, fields []string) (statementTimeseriesResponse, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	types := prefixedStatementFields(fields, frequency)
	params.Set("type", strings.Join(types, ","))
	params.Set("period1", "0")
	params.Set("period2", fmt.Sprintf("%d", time.Now().UTC().Unix()))
	params.Set("merge", "false")
	params.Set("padTimeSeries", "true")
	params.Set("lang", "en-US")
	params.Set("region", "US")

	var resp statementTimeseriesResponse
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.timeseriesBase + url.PathEscape(symbol),
		Params:   params,
		CacheKey: "fundamentals-timeseries:" + key + ":v4:" + symbol,
		TTL:      30 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return statementTimeseriesResponse{}, err
	}
	return resp, nil
}

func mergeStatementTimeseriesResponses(items ...statementTimeseriesResponse) statementTimeseriesResponse {
	var merged statementTimeseriesResponse
	for _, item := range items {
		merged.Timeseries.Result = append(merged.Timeseries.Result, item.Timeseries.Result...)
	}
	return merged
}

func deriveSupplementalROIC(resp statementTimeseriesResponse) (float64, float64, bool) {
	if investedCapital, roic, ok := deriveQuarterlyTTMROIC(resp); ok {
		return investedCapital, roic, true
	}
	return deriveAnnualROIC(resp)
}

func deriveSupplementalProfitability(resp statementTimeseriesResponse) (supplementalProfitability, bool) {
	revenue, ok := trailingStatementValue(resp, "TotalRevenue")
	if !ok || revenue <= 0 {
		return supplementalProfitability{}, false
	}

	grossProfit, grossOK := trailingStatementValue(resp, "GrossProfit")
	operatingIncome, opOK := trailingStatementValue(resp, "OperatingIncome")
	netIncome, netOK := trailingStatementValue(resp, "NetIncome")
	assetsSeries := statementSeriesWithAnnualFallback(resp, "TotalAssets")
	equitySeries := statementSeriesWithAnnualFallback(resp, "StockholdersEquity")

	metrics := supplementalProfitability{}
	if grossOK {
		metrics.grossMargin = grossProfit / revenue
	}
	if opOK {
		metrics.operatingMargin = operatingIncome / revenue
	}
	if netOK {
		metrics.profitMargin = netIncome / revenue
	}
	if netOK {
		if avgAssets, ok := averageBalanceBase(assetsSeries); ok && avgAssets > 0 {
			metrics.returnOnAssets = netIncome / avgAssets
		}
		if avgEquity, ok := averageBalanceBase(equitySeries); ok && avgEquity > 0 {
			metrics.returnOnEquity = netIncome / avgEquity
		}
	}
	if opOK {
		if investedCapital, roic, ok := deriveSupplementalROIC(resp); ok {
			metrics.investedCapital = investedCapital
			metrics.returnOnIC = roic
		}
	}

	return metrics, true
}

func deriveQuarterlyTTMROIC(resp statementTimeseriesResponse) (float64, float64, bool) {
	operatingIncome, ok := sumStatementValues(resp, domain.StatementFrequencyQuarterly, 4, "OperatingIncome", "EBIT")
	if !ok {
		return 0, 0, false
	}

	taxRate, ok := effectiveTaxRateTTM(resp, domain.StatementFrequencyQuarterly)
	if !ok {
		return 0, 0, false
	}

	series := investedCapitalSeries(resp, domain.StatementFrequencyQuarterly)
	if len(series) == 0 {
		return 0, 0, false
	}

	currentInvestedCapital := series[0].value
	if currentInvestedCapital <= 0 {
		return 0, 0, false
	}

	averageInvestedCapital := currentInvestedCapital
	if yearAgoIdx := comparisonIndex(series, 0, 45*24*time.Hour); yearAgoIdx > 0 && series[yearAgoIdx].value > 0 {
		averageInvestedCapital = (currentInvestedCapital + series[yearAgoIdx].value) / 2
	}
	if averageInvestedCapital <= 0 {
		return 0, 0, false
	}

	beginningInvestedCapital := averageInvestedCapital
	if yearAgoIdx := comparisonIndex(series, 0, 45*24*time.Hour); yearAgoIdx > 0 && series[yearAgoIdx].value > 0 {
		beginningInvestedCapital = series[yearAgoIdx].value
	}

	nopat := operatingIncome * (1 - taxRate)
	return currentInvestedCapital, nopat / beginningInvestedCapital, true
}

func deriveAnnualROIC(resp statementTimeseriesResponse) (float64, float64, bool) {
	operatingIncome, ok := firstPresentStatementValue(resp, domain.StatementFrequencyAnnual, 0, "OperatingIncome", "EBIT")
	if !ok {
		return 0, 0, false
	}

	taxRate, ok := effectiveTaxRateTTM(resp, domain.StatementFrequencyAnnual)
	if !ok {
		return 0, 0, false
	}

	series := investedCapitalSeries(resp, domain.StatementFrequencyAnnual)
	if len(series) == 0 || series[0].value <= 0 {
		return 0, 0, false
	}

	averageInvestedCapital := series[0].value
	if len(series) > 1 && series[1].value > 0 {
		averageInvestedCapital = (series[0].value + series[1].value) / 2
	}
	if averageInvestedCapital <= 0 {
		return 0, 0, false
	}

	beginningInvestedCapital := averageInvestedCapital
	if len(series) > 1 && series[1].value > 0 {
		beginningInvestedCapital = series[1].value
	}

	nopat := operatingIncome * (1 - taxRate)
	return series[0].value, nopat / beginningInvestedCapital, true
}

func effectiveTaxRateTTM(resp statementTimeseriesResponse, frequency domain.StatementFrequency) (float64, bool) {
	count := 1
	if frequency == domain.StatementFrequencyQuarterly {
		count = 4
	}

	taxProvision, ok := sumStatementValues(resp, frequency, count, "TaxProvision")
	if !ok {
		return 0, false
	}
	pretaxIncome, ok := sumStatementValues(resp, frequency, count, "PretaxIncome")
	if !ok || pretaxIncome == 0 {
		return 0, false
	}

	taxRate := taxProvision / pretaxIncome
	if math.IsNaN(taxRate) || math.IsInf(taxRate, 0) {
		return 0, false
	}
	if taxRate < 0 {
		taxRate = 0
	}
	if taxRate > 1 {
		taxRate = 1
	}
	return taxRate, true
}

func investedCapitalSeries(resp statementTimeseriesResponse, frequency domain.StatementFrequency) []datedFundamentalsValue {
	if values := statementValues(resp, frequency, "InvestedCapital"); len(values) > 0 {
		return values
	}

	equitySeries := statementValues(resp, frequency, "StockholdersEquity")
	debtSeries := statementValues(resp, frequency, "TotalDebt")
	cashSeries := statementValues(resp, frequency, "CashCashEquivalentsAndShortTermInvestments")
	if len(cashSeries) == 0 {
		cashSeries = statementValues(resp, frequency, "CashAndCashEquivalents")
	}

	if len(equitySeries) == 0 && len(debtSeries) == 0 {
		return nil
	}

	byDate := make(map[string]datedFundamentalsValue, len(equitySeries)+len(debtSeries))
	for _, item := range equitySeries {
		entry := byDate[item.date]
		entry.date = item.date
		entry.end = item.end
		entry.value += item.value
		byDate[item.date] = entry
	}
	for _, item := range debtSeries {
		entry := byDate[item.date]
		entry.date = item.date
		entry.end = item.end
		entry.value += item.value
		byDate[item.date] = entry
	}
	for _, item := range cashSeries {
		entry := byDate[item.date]
		entry.date = item.date
		entry.end = item.end
		entry.value -= item.value
		byDate[item.date] = entry
	}

	if len(byDate) == 0 {
		return nil
	}

	values := make([]datedFundamentalsValue, 0, len(byDate))
	for _, item := range byDate {
		if item.value > 0 {
			values = append(values, item)
		}
	}
	sort.Slice(values, func(i, j int) bool {
		return values[i].date > values[j].date
	})
	return values
}

func firstPresentStatementValue(resp statementTimeseriesResponse, frequency domain.StatementFrequency, idx int, keys ...string) (float64, bool) {
	for _, key := range keys {
		if value, ok := nthStatementValue(resp, frequency, key, idx); ok {
			return value, true
		}
	}
	return 0, false
}

func nthStatementValue(resp statementTimeseriesResponse, frequency domain.StatementFrequency, key string, idx int) (float64, bool) {
	values := statementValues(resp, frequency, key)
	if idx < 0 || idx >= len(values) {
		return 0, false
	}
	return values[idx].value, true
}

func sumStatementValues(resp statementTimeseriesResponse, frequency domain.StatementFrequency, count int, keys ...string) (float64, bool) {
	for _, key := range keys {
		values := statementValues(resp, frequency, key)
		if len(values) < count {
			continue
		}
		sum := 0.0
		for i := 0; i < count; i++ {
			sum += values[i].value
		}
		return sum, true
	}
	return 0, false
}

func comparisonIndex(values []datedFundamentalsValue, idx int, tolerance time.Duration) int {
	if idx < 0 || idx >= len(values) || values[idx].end.IsZero() {
		return -1
	}
	target := values[idx].end.AddDate(-1, 0, 0)
	bestIdx := -1
	bestDelta := time.Duration(1<<63 - 1)
	for candidate := idx + 1; candidate < len(values); candidate++ {
		if values[candidate].end.IsZero() {
			continue
		}
		delta := values[candidate].end.Sub(target)
		if delta < 0 {
			delta = -delta
		}
		if delta <= tolerance && delta < bestDelta {
			bestIdx = candidate
			bestDelta = delta
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	if idx+4 < len(values) {
		return idx + 4
	}
	return -1
}

func statementValues(resp statementTimeseriesResponse, frequency domain.StatementFrequency, key string) []datedFundamentalsValue {
	byDate := make(map[string]float64)
	for _, result := range resp.Timeseries.Result {
		fullKey := firstStatementKey(result)
		if fullKey == "" || fullKey != string(frequency)+key {
			continue
		}
		points := result.Data[fullKey]
		for _, point := range points {
			date := strings.TrimSpace(point.AsOfDate)
			if date == "" || !point.ReportedValue.Present {
				continue
			}
			byDate[date] = point.ReportedValue.Value
		}
	}
	if len(byDate) == 0 {
		return nil
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] > dates[j]
	})

	values := make([]datedFundamentalsValue, 0, len(dates))
	for _, date := range dates {
		end, _ := time.Parse("2006-01-02", date)
		values = append(values, datedFundamentalsValue{
			date:  date,
			end:   end,
			value: byDate[date],
		})
	}
	return values
}

func trailingStatementValue(resp statementTimeseriesResponse, key string) (float64, bool) {
	if value, ok := sumStatementValues(resp, domain.StatementFrequencyQuarterly, 4, key); ok {
		return value, true
	}
	return firstPresentStatementValue(resp, domain.StatementFrequencyAnnual, 0, key)
}

func statementSeriesWithAnnualFallback(resp statementTimeseriesResponse, key string) []datedFundamentalsValue {
	if values := statementValues(resp, domain.StatementFrequencyQuarterly, key); len(values) > 0 {
		return values
	}
	return statementValues(resp, domain.StatementFrequencyAnnual, key)
}

func averageBalanceBase(series []datedFundamentalsValue) (float64, bool) {
	if len(series) == 0 || series[0].value <= 0 {
		return 0, false
	}
	base := series[0].value
	if yearAgoIdx := comparisonIndex(series, 0, 45*24*time.Hour); yearAgoIdx > 0 && series[yearAgoIdx].value > 0 {
		return (base + series[yearAgoIdx].value) / 2, true
	}
	if len(series) > 1 && series[1].value > 0 {
		return (base + series[1].value) / 2, true
	}
	return base, true
}

func deriveSupplementalCashFlow(resp statementTimeseriesResponse) (supplementalCashFlow, bool) {
	operatingCashFlow, ok := trailingCashFlowValueWithFallback(resp, "OperatingCashFlow", "CashFlowFromContinuingOperatingActivities")
	if !ok {
		return supplementalCashFlow{}, false
	}

	capitalExpenditure, capexOK := trailingCashFlowValueWithFallback(resp, "CapitalExpenditure")
	if capexOK {
		return supplementalCashFlow{
			operatingCashFlow: operatingCashFlow,
			freeCashFlow:      computeStatementFreeCashFlow(operatingCashFlow, capitalExpenditure),
		}, true
	}

	freeCashFlow, fcfOK := trailingCashFlowValueWithFallback(resp, "FreeCashFlow")
	if !fcfOK {
		return supplementalCashFlow{operatingCashFlow: operatingCashFlow}, true
	}

	return supplementalCashFlow{
		operatingCashFlow: operatingCashFlow,
		freeCashFlow:      freeCashFlow,
	}, true
}

func trailingCashFlowValueWithFallback(resp statementTimeseriesResponse, keys ...string) (float64, bool) {
	quarterly := statementSeriesForKeys(resp, domain.StatementFrequencyQuarterly, keys...)
	annual := statementSeriesForKeys(resp, domain.StatementFrequencyAnnual, keys...)

	if len(annual) > 0 && len(quarterly) > 0 && annual[0].date == quarterly[0].date {
		return annual[0].value, true
	}

	if len(quarterly) >= 4 {
		sum := 0.0
		for i := 0; i < 4; i++ {
			sum += quarterly[i].value
		}
		return sum, true
	}

	if len(annual) > 0 {
		return annual[0].value, true
	}
	if len(quarterly) > 0 {
		return quarterly[0].value, true
	}
	return 0, false
}

func statementSeriesForKeys(resp statementTimeseriesResponse, frequency domain.StatementFrequency, keys ...string) []datedFundamentalsValue {
	for _, key := range keys {
		if values := statementValues(resp, frequency, key); len(values) > 0 {
			return values
		}
	}
	return nil
}

func trailingStatementValueWithFallback(resp statementTimeseriesResponse, keys ...string) (float64, bool) {
	if value, ok := sumStatementValues(resp, domain.StatementFrequencyQuarterly, 4, keys...); ok {
		return value, true
	}
	return firstPresentStatementValue(resp, domain.StatementFrequencyAnnual, 0, keys...)
}

func computeStatementFreeCashFlow(operatingCashFlow, capitalExpenditure float64) float64 {
	if capitalExpenditure > 0 {
		return operatingCashFlow - capitalExpenditure
	}
	return operatingCashFlow + capitalExpenditure
}
