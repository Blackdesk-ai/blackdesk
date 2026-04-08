package yahoo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"blackdesk/internal/domain"
)

var incomeStatementFields = []string{
	"TotalRevenue",
	"OperatingRevenue",
	"CostOfRevenue",
	"GrossProfit",
	"OperatingExpense",
	"SellingGeneralAndAdministration",
	"ResearchAndDevelopment",
	"OperatingIncome",
	"InterestExpense",
	"InterestIncome",
	"PretaxIncome",
	"TaxProvision",
	"NetIncome",
	"DilutedEPS",
	"BasicEPS",
	"DilutedAverageShares",
	"BasicAverageShares",
	"EBIT",
	"EBITDA",
	"TotalExpenses",
	"DepreciationAndAmortization",
}

var balanceSheetFields = []string{
	"TotalAssets",
	"CurrentAssets",
	"CashCashEquivalentsAndShortTermInvestments",
	"CashAndCashEquivalents",
	"AccountsReceivable",
	"Inventory",
	"OtherCurrentAssets",
	"TotalNonCurrentAssets",
	"NetPPE",
	"GrossPPE",
	"AccumulatedDepreciation",
	"Goodwill",
	"GoodwillAndOtherIntangibleAssets",
	"InvestmentsAndAdvances",
	"TotalLiabilitiesNetMinorityInterest",
	"CurrentLiabilities",
	"AccountsPayable",
	"CurrentDebt",
	"OtherCurrentLiabilities",
	"LongTermDebt",
	"LongTermDebtAndCapitalLeaseObligation",
	"OtherNonCurrentLiabilities",
	"StockholdersEquity",
	"CommonStockEquity",
	"RetainedEarnings",
	"WorkingCapital",
	"InvestedCapital",
	"TotalDebt",
	"NetDebt",
	"OrdinarySharesNumber",
}

var cashFlowFields = []string{
	"OperatingCashFlow",
	"CashFlowFromContinuingOperatingActivities",
	"NetIncomeFromContinuingOperations",
	"DepreciationAndAmortization",
	"DeferredIncomeTax",
	"ChangeInWorkingCapital",
	"ChangeInReceivables",
	"ChangeInInventory",
	"ChangeInAccountPayable",
	"StockBasedCompensation",
	"OtherNonCashItems",
	"InvestingCashFlow",
	"CashFlowFromContinuingInvestingActivities",
	"CapitalExpenditure",
	"NetInvestmentPurchaseAndSale",
	"FinancingCashFlow",
	"CashFlowFromContinuingFinancingActivities",
	"NetIssuancePaymentsOfDebt",
	"NetCommonStockIssuance",
	"RepurchaseOfCapitalStock",
	"CashDividendsPaid",
	"EndCashPosition",
	"BeginningCashPosition",
	"ChangesinCash",
	"FreeCashFlow",
}

type statementTimeseriesResponse struct {
	Timeseries struct {
		Result []statementTimeseriesResult `json:"result"`
		Error  any                         `json:"error"`
	} `json:"timeseries"`
}

type statementTimeseriesResult struct {
	Meta statementMeta
	Data map[string][]statementTimeseriesPoint
}

type statementMeta struct {
	Type []string `json:"type"`
}

type statementTimeseriesPoint struct {
	AsOfDate      string                 `json:"asOfDate"`
	CurrencyCode  string                 `json:"currencyCode"`
	PeriodType    string                 `json:"periodType"`
	ReportedValue statementReportedValue `json:"reportedValue"`
}

type statementReportedValue struct {
	Value   float64
	Present bool
}

func (r *statementTimeseriesResult) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if metaRaw, ok := raw["meta"]; ok {
		if err := json.Unmarshal(metaRaw, &r.Meta); err != nil {
			return err
		}
		delete(raw, "meta")
	}

	r.Data = make(map[string][]statementTimeseriesPoint, len(raw))
	for key, rawValue := range raw {
		var points []statementTimeseriesPoint
		if err := json.Unmarshal(rawValue, &points); err != nil {
			continue
		}
		r.Data[key] = points
	}
	return nil
}

func (v *statementReportedValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var direct float64
	if err := json.Unmarshal(data, &direct); err == nil {
		v.Value = direct
		v.Present = true
		return nil
	}

	var wrapped struct {
		Raw json.RawMessage `json:"raw"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if len(wrapped.Raw) == 0 || string(wrapped.Raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(wrapped.Raw, &direct); err == nil {
		v.Value = direct
		v.Present = true
		return nil
	}

	var nested struct {
		ParsedValue *float64 `json:"parsedValue"`
	}
	if err := json.Unmarshal(wrapped.Raw, &nested); err == nil && nested.ParsedValue != nil {
		v.Value = *nested.ParsedValue
		v.Present = true
		return nil
	}

	return nil
}

func (p *Provider) GetStatement(ctx context.Context, symbol string, kind domain.StatementKind, frequency domain.StatementFrequency) (domain.FinancialStatement, error) {
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if normalizedSymbol == "" {
		return domain.FinancialStatement{}, fmt.Errorf("statement symbol is required")
	}

	fields, err := statementFields(kind)
	if err != nil {
		return domain.FinancialStatement{}, err
	}

	now := time.Now().UTC()
	params := url.Values{}
	params.Set("symbol", normalizedSymbol)
	params.Set("type", strings.Join(prefixedStatementFields(fields, frequency), ","))
	params.Set("period1", fmt.Sprintf("%d", now.AddDate(-10, 0, 0).Unix()))
	params.Set("period2", fmt.Sprintf("%d", now.Unix()))
	params.Set("merge", "false")
	params.Set("padTimeSeries", "true")
	params.Set("lang", "en-US")
	params.Set("region", "US")

	var resp statementTimeseriesResponse
	err = p.fetchJSON(ctx, requestSpec{
		URL:      p.timeseriesBase + url.PathEscape(normalizedSymbol),
		Params:   params,
		CacheKey: fmt.Sprintf("statement:%s:%s:%s", normalizedSymbol, kind, frequency),
		TTL:      30 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return domain.FinancialStatement{}, err
	}

	return normalizeFinancialStatement(normalizedSymbol, kind, frequency, fields, resp)
}

func statementFields(kind domain.StatementKind) ([]string, error) {
	switch kind {
	case domain.StatementKindIncome:
		return incomeStatementFields, nil
	case domain.StatementKindBalanceSheet:
		return balanceSheetFields, nil
	case domain.StatementKindCashFlow:
		return cashFlowFields, nil
	default:
		return nil, fmt.Errorf("unsupported statement kind %q", kind)
	}
}

func prefixedStatementFields(fields []string, frequency domain.StatementFrequency) []string {
	prefix := string(frequency)
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		out = append(out, prefix+field)
	}
	return out
}

func normalizeFinancialStatement(symbol string, kind domain.StatementKind, frequency domain.StatementFrequency, fields []string, resp statementTimeseriesResponse) (domain.FinancialStatement, error) {
	if len(resp.Timeseries.Result) == 0 {
		return domain.FinancialStatement{}, fmt.Errorf("statement not found")
	}

	rowByKey := make(map[string]map[string]domain.StatementValue, len(fields))
	periodByDate := make(map[string]domain.StatementPeriod)
	currency := ""

	for _, result := range resp.Timeseries.Result {
		fullKey := firstStatementKey(result)
		if fullKey == "" {
			continue
		}

		key := stripStatementPrefix(fullKey)
		points := result.Data[fullKey]
		if len(points) == 0 {
			continue
		}
		if rowByKey[key] == nil {
			rowByKey[key] = make(map[string]domain.StatementValue, len(points))
		}

		for _, point := range points {
			date := strings.TrimSpace(point.AsOfDate)
			if date == "" || !point.ReportedValue.Present {
				continue
			}
			rowByKey[key][date] = domain.StatementValue{
				Value:   point.ReportedValue.Value,
				Present: true,
			}
			if currency == "" && strings.TrimSpace(point.CurrencyCode) != "" {
				currency = strings.TrimSpace(point.CurrencyCode)
			}
			if _, ok := periodByDate[date]; !ok {
				periodByDate[date] = buildStatementPeriod(date, frequency)
			}
		}
	}

	if len(periodByDate) == 0 {
		return domain.FinancialStatement{}, fmt.Errorf("statement periods not found")
	}

	dates := make([]string, 0, len(periodByDate))
	for date := range periodByDate {
		dates = append(dates, date)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] > dates[j]
	})

	periods := make([]domain.StatementPeriod, 0, len(dates))
	for _, date := range dates {
		periods = append(periods, periodByDate[date])
	}

	rows := make([]domain.StatementRow, 0, len(fields))
	for _, key := range fields {
		dateValues, ok := rowByKey[key]
		if !ok {
			continue
		}
		values := make([]domain.StatementValue, len(dates))
		anyPresent := false
		for i, date := range dates {
			if value, ok := dateValues[date]; ok {
				values[i] = value
				anyPresent = true
			}
		}
		if !anyPresent {
			continue
		}
		rows = append(rows, domain.StatementRow{
			Key:    key,
			Label:  humanizeStatementKey(key),
			Values: values,
		})
	}

	if len(rows) == 0 {
		return domain.FinancialStatement{}, fmt.Errorf("statement rows not found")
	}

	return domain.FinancialStatement{
		Symbol:    symbol,
		Kind:      kind,
		Frequency: frequency,
		Currency:  currency,
		Periods:   periods,
		Rows:      rows,
		Freshness: domain.FreshnessLive,
		Provider:  "yahoo",
		UpdatedAt: time.Now(),
	}, nil
}

func firstStatementKey(result statementTimeseriesResult) string {
	if len(result.Meta.Type) > 0 && strings.TrimSpace(result.Meta.Type[0]) != "" {
		return strings.TrimSpace(result.Meta.Type[0])
	}
	for key := range result.Data {
		return key
	}
	return ""
}

func stripStatementPrefix(key string) string {
	for _, prefix := range []string{"annual", "quarterly", "trailing"} {
		if stripped, ok := strings.CutPrefix(key, prefix); ok {
			return stripped
		}
	}
	return key
}

func buildStatementPeriod(date string, frequency domain.StatementFrequency) domain.StatementPeriod {
	period := domain.StatementPeriod{Label: date}
	if ts, err := time.Parse("2006-01-02", date); err == nil {
		period.EndDate = ts
		period.FiscalYear = ts.Year()
		if frequency == domain.StatementFrequencyAnnual {
			period.Label = fmt.Sprintf("FY %d", ts.Year())
		}
	}
	return period
}

func humanizeStatementKey(key string) string {
	if key == "" {
		return ""
	}

	runes := []rune(key)
	out := make([]rune, 0, len(runes)+4)
	for i, r := range runes {
		if i > 0 {
			prev := runes[i-1]
			var next rune
			if i+1 < len(runes) {
				next = runes[i+1]
			}
			if unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && next != 0 && unicode.IsLower(next))) {
				out = append(out, ' ')
			}
		}
		out = append(out, r)
	}
	return string(out)
}
