package tui

import (
	"context"
	"strings"
	"time"

	"blackdesk/internal/application"
	"blackdesk/internal/domain"
)

type testProvider struct{}

type countingHistoryProvider struct {
	historyCalls []string
}

type aiPrepProvider struct {
	quoteCalls        int
	quotesCalls       int
	historyCalls      []string
	statementCalls    []string
	newsCalls         int
	fundamentalsCalls int
}

type statementsProvider struct {
	testProvider
}

type insidersProvider struct {
	testProvider
}

type researchProvider struct {
	testProvider
}

type marketNewsProvider struct {
	testProvider
	marketNewsCalls int
	quotesCalls     int
}

type marketRiskProvider struct {
	data domain.MarketRiskSnapshot
	err  error
}

type filingsProvider struct{}

func (testProvider) Name() string { return "test" }

func (testProvider) Capabilities() domain.ProviderCapabilities { return domain.ProviderCapabilities{} }

func (testProvider) GetQuote(context.Context, string) (domain.QuoteSnapshot, error) {
	return domain.QuoteSnapshot{}, nil
}

func (testProvider) GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error) {
	return nil, nil
}

func (testProvider) GetHistory(context.Context, string, string, string) (domain.PriceSeries, error) {
	return domain.PriceSeries{}, nil
}

func (testProvider) GetNews(context.Context, string) ([]domain.NewsItem, error) {
	return nil, nil
}

func (testProvider) GetFundamentals(context.Context, string) (domain.FundamentalsSnapshot, error) {
	return domain.FundamentalsSnapshot{}, nil
}

func (testProvider) GetEarnings(context.Context, string) (domain.EarningsSnapshot, error) {
	return sampleEarningsSnapshot(), nil
}

func (testProvider) SearchSymbols(context.Context, string) ([]domain.SymbolRef, error) {
	return nil, nil
}

func (p *marketNewsProvider) GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error) {
	p.quotesCalls++
	return nil, nil
}

func (p *marketNewsProvider) GetMarketNews(context.Context) ([]domain.NewsItem, error) {
	p.marketNewsCalls++
	return nil, nil
}

func (p marketRiskProvider) GetMarketRisk(context.Context) (domain.MarketRiskSnapshot, error) {
	if p.err != nil {
		return domain.MarketRiskSnapshot{}, p.err
	}
	return p.data, nil
}

var _ application.MarketRiskProvider = marketRiskProvider{}
var _ application.FilingsProvider = filingsProvider{}

func (filingsProvider) GetFilings(context.Context, string) (domain.FilingsSnapshot, error) {
	return domain.FilingsSnapshot{
		Symbol:      "AAPL",
		CompanyName: "Apple Inc.",
		CIK:         "0000320193",
		Items: []domain.FilingItem{
			{
				AccessionNumber:       "0000320193-24-000123",
				Form:                  "10-K",
				FilingDate:            time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
				PrimaryDocument:       "aapl-20240928x10k.htm",
				PrimaryDocDescription: "Annual report",
				URL:                   "https://www.sec.gov/Archives/edgar/data/320193/000032019324000123/aapl-20240928x10k.htm",
				IsXBRL:                true,
				IsInlineXBRL:          true,
			},
			{
				AccessionNumber:       "0000320193-24-000098",
				Form:                  "8-K",
				FilingDate:            time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
				PrimaryDocument:       "aapl-8k.htm",
				PrimaryDocDescription: "Current report",
				URL:                   "https://www.sec.gov/Archives/edgar/data/320193/000032019324000098/aapl-8k.htm",
			},
		},
		Freshness: domain.FreshnessLive,
		Provider:  "sec",
		UpdatedAt: time.Now(),
	}, nil
}

func (filingsProvider) GetFilingDocument(_ context.Context, item domain.FilingItem) (domain.FilingDocument, error) {
	return domain.FilingDocument{
		Item:        item,
		ContentType: "text/html",
		Text:        "Item 1. Business\nRevenue grew 12% year over year.\nRisk factors include supply chain concentration.\nManagement highlighted services margin expansion.",
		Provider:    "sec",
		RetrievedAt: time.Now(),
	}, nil
}

func (p *countingHistoryProvider) Name() string { return "test" }

func (p *countingHistoryProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{}
}

func (p *countingHistoryProvider) GetQuote(context.Context, string) (domain.QuoteSnapshot, error) {
	return domain.QuoteSnapshot{}, nil
}

func (p *countingHistoryProvider) GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error) {
	return nil, nil
}

func (p *countingHistoryProvider) GetHistory(_ context.Context, symbol, rangeKey, interval string) (domain.PriceSeries, error) {
	p.historyCalls = append(p.historyCalls, symbol+"|"+rangeKey+"|"+interval)
	candles := make([]domain.Candle, 0, 260)
	for i := 260; i >= 0; i-- {
		price := 100.0 + float64(260-i)*0.5
		candles = append(candles, domain.Candle{
			Time:   time.Now().AddDate(0, 0, -i),
			Open:   price - 1,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1_000_000,
		})
	}
	return domain.PriceSeries{Symbol: symbol, Range: rangeKey, Interval: interval, Candles: candles}, nil
}

func (p *countingHistoryProvider) GetNews(context.Context, string) ([]domain.NewsItem, error) {
	return nil, nil
}

func (p *countingHistoryProvider) GetFundamentals(context.Context, string) (domain.FundamentalsSnapshot, error) {
	return domain.FundamentalsSnapshot{}, nil
}

func (p *countingHistoryProvider) GetEarnings(context.Context, string) (domain.EarningsSnapshot, error) {
	return sampleEarningsSnapshot(), nil
}

func (p *countingHistoryProvider) SearchSymbols(context.Context, string) ([]domain.SymbolRef, error) {
	return nil, nil
}

func (p *aiPrepProvider) Name() string { return "test" }

func (p *aiPrepProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{Statements: true}
}

func (p *aiPrepProvider) GetQuote(context.Context, string) (domain.QuoteSnapshot, error) {
	p.quoteCalls++
	return domain.QuoteSnapshot{Symbol: "AAPL", Price: 210}, nil
}

func (p *aiPrepProvider) GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error) {
	p.quotesCalls++
	return []domain.QuoteSnapshot{{Symbol: "SPY", Price: 500}}, nil
}

func (p *aiPrepProvider) GetHistory(_ context.Context, symbol, rangeKey, interval string) (domain.PriceSeries, error) {
	p.historyCalls = append(p.historyCalls, symbol+"|"+rangeKey+"|"+interval)
	candles := make([]domain.Candle, 0, 260)
	for i := 260; i >= 0; i-- {
		price := 100.0 + float64(260-i)*0.5
		candles = append(candles, domain.Candle{
			Time:   time.Now().AddDate(0, 0, -i),
			Open:   price - 1,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1_000_000,
		})
	}
	return domain.PriceSeries{Symbol: symbol, Range: rangeKey, Interval: interval, Candles: candles}, nil
}

func (p *aiPrepProvider) GetStatement(_ context.Context, symbol string, kind domain.StatementKind, freq domain.StatementFrequency) (domain.FinancialStatement, error) {
	p.statementCalls = append(p.statementCalls, symbol+"|"+string(kind)+"|"+string(freq))
	return domain.FinancialStatement{
		Symbol:    symbol,
		Kind:      kind,
		Frequency: freq,
		Periods:   []domain.StatementPeriod{{Label: "FY 2024"}},
		Rows: []domain.StatementRow{
			{
				Key:   "TotalRevenue",
				Label: "Total Revenue",
				Values: []domain.StatementValue{
					{Value: 391_035_000_000, Present: true},
				},
			},
		},
	}, nil
}

func (p *aiPrepProvider) GetNews(context.Context, string) ([]domain.NewsItem, error) {
	p.newsCalls++
	return []domain.NewsItem{{Title: "News"}}, nil
}

func (p *aiPrepProvider) GetFundamentals(context.Context, string) (domain.FundamentalsSnapshot, error) {
	p.fundamentalsCalls++
	return domain.FundamentalsSnapshot{Symbol: "AAPL", Description: "Apple", MarketCap: 1}, nil
}

func (p *aiPrepProvider) GetEarnings(context.Context, string) (domain.EarningsSnapshot, error) {
	return sampleEarningsSnapshot(), nil
}

func (p *aiPrepProvider) SearchSymbols(context.Context, string) ([]domain.SymbolRef, error) {
	return nil, nil
}

func (statementsProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{Statements: true}
}

func (insidersProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{Insiders: true}
}

func (researchProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{Statements: true, Insiders: true}
}

func sampleEarningsSnapshot() domain.EarningsSnapshot {
	return domain.EarningsSnapshot{
		Symbol:      "AAPL",
		CompanyName: "Apple Inc.",
		Items: []domain.EarningsItem{
			{
				Kind:           "upcoming",
				Title:          "Next earnings",
				WindowStart:    time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
				WindowEnd:      time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
				EPSEstimate:    1.62,
				EPSLow:         1.55,
				EPSHigh:        1.71,
				RevenueAverage: 95_400_000_000,
				RevenueLow:     94_100_000_000,
				RevenueHigh:    97_800_000_000,
			},
			{
				Kind:            "reported",
				Title:           "Reported quarter",
				QuarterEnd:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
				EPSEstimate:     1.48,
				EPSActual:       1.52,
				EPSDifference:   0.04,
				SurprisePercent: 0.027,
			},
			{
				Kind:            "reported",
				Title:           "Reported quarter",
				QuarterEnd:      time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
				EPSEstimate:     1.41,
				EPSActual:       1.37,
				EPSDifference:   -0.04,
				SurprisePercent: -0.028,
			},
		},
		Estimates: []domain.EarningsEstimate{
			{Period: "0q", EPSAverage: 1.62, RevenueAverage: 95_400_000_000},
			{Period: "+1q", EPSAverage: 1.77, RevenueAverage: 102_300_000_000},
			{Period: "0y", EPSAverage: 7.08, RevenueAverage: 403_000_000_000},
			{Period: "+1y", EPSAverage: 7.44, RevenueAverage: 420_000_000_000},
		},
		Freshness: domain.FreshnessLive,
		Provider:  "test",
		UpdatedAt: time.Now(),
	}
}

func (statementsProvider) GetStatement(_ context.Context, _ string, kind domain.StatementKind, freq domain.StatementFrequency) (domain.FinancialStatement, error) {
	labelA := "FY 2024"
	labelB := "FY 2023"
	if freq == domain.StatementFrequencyQuarterly {
		labelA = "2024-09-30"
		labelB = "2024-06-30"
	}
	rowLabel := "Total Revenue"
	rowKey := "TotalRevenue"
	switch kind {
	case domain.StatementKindBalanceSheet:
		rowLabel = "Total Assets"
		rowKey = "TotalAssets"
	case domain.StatementKindCashFlow:
		rowLabel = "Operating Cash Flow"
		rowKey = "OperatingCashFlow"
	}
	return domain.FinancialStatement{
		Symbol:    "AAPL",
		Kind:      kind,
		Frequency: freq,
		Periods: []domain.StatementPeriod{
			{Label: labelA},
			{Label: labelB},
		},
		Rows: []domain.StatementRow{
			{
				Key:   rowKey,
				Label: rowLabel,
				Values: []domain.StatementValue{
					{Value: 391_035_000_000, Present: true},
					{Value: 383_285_000_000, Present: true},
				},
			},
			{
				Key:   "NetIncome",
				Label: "Net Income",
				Values: []domain.StatementValue{
					{Value: 100_913_000_000, Present: true},
					{Value: 96_995_000_000, Present: true},
				},
			},
		},
	}, nil
}

func (insidersProvider) GetInsiders(_ context.Context, symbol string) (domain.InsiderSnapshot, error) {
	return sampleInsiderSnapshot(strings.ToUpper(symbol)), nil
}

func (researchProvider) GetStatement(_ context.Context, _ string, kind domain.StatementKind, freq domain.StatementFrequency) (domain.FinancialStatement, error) {
	return statementsProvider{}.GetStatement(context.Background(), "AAPL", kind, freq)
}

func (researchProvider) GetInsiders(_ context.Context, symbol string) (domain.InsiderSnapshot, error) {
	return sampleInsiderSnapshot(strings.ToUpper(symbol)), nil
}

func sampleInsiderSnapshot(symbol string) domain.InsiderSnapshot {
	return domain.InsiderSnapshot{
		Symbol: symbol,
		PurchaseActivity: domain.InsiderPurchaseActivity{
			Period:                  "6m",
			BuyShares:               305_864,
			BuyTransactions:         3,
			SellShares:              1_606_142,
			SellTransactions:        12,
			NetShares:               -1_300_278,
			NetTransactions:         15,
			TotalInsiderShares:      2_100_000,
			NetPercentInsiderShares: -0.382,
		},
		Transactions: []domain.InsiderTransaction{
			{
				Insider:   "SIEFFERT KRISTEN N",
				Relation:  "President",
				Ownership: "Direct",
				Action:    "Sale",
				Date:      time.Unix(1769904000, 0),
				Shares:    750,
				Value:     17_490,
				Text:      "Sale at price 23.32 per share.",
			},
			{
				Insider:   "THORNOCK TAI A",
				Relation:  "Officer",
				Ownership: "Indirect",
				Action:    "Buy",
				Date:      time.Unix(1736985600, 0),
				Shares:    1_100,
				Value:     27_038,
				Text:      "Purchase at price 24.58 per share.",
			},
		},
		Roster: []domain.InsiderRosterMember{
			{
				Name:                "SIEFFERT KRISTEN N",
				Relation:            "President",
				LatestTransaction:   "Sale",
				LatestTransactionAt: time.Unix(1769904000, 0),
				SharesOwnedDirectly: 2_100_000,
			},
		},
	}
}

func populateAllStatementCache(m *Model, symbol string) {
	for _, req := range aiStatementRequests {
		periodLabel := "FY 2024"
		prevLabel := "FY 2023"
		if req.frequency == domain.StatementFrequencyQuarterly {
			periodLabel = "2024-09-30"
			prevLabel = "2024-06-30"
		}
		stmt := domain.FinancialStatement{
			Symbol:    symbol,
			Kind:      req.kind,
			Frequency: req.frequency,
			Periods: []domain.StatementPeriod{
				{Label: periodLabel},
				{Label: prevLabel},
			},
		}
		switch req.kind {
		case domain.StatementKindIncome:
			stmt.Rows = []domain.StatementRow{
				{
					Key:   "TotalRevenue",
					Label: "Total Revenue",
					Values: []domain.StatementValue{
						{Value: 391_035_000_000, Present: true},
						{Value: 383_285_000_000, Present: true},
					},
				},
				{
					Key:   "NetIncome",
					Label: "Net Income",
					Values: []domain.StatementValue{
						{Value: 100_913_000_000, Present: true},
						{Value: 96_995_000_000, Present: true},
					},
				},
			}
		case domain.StatementKindBalanceSheet:
			stmt.Rows = []domain.StatementRow{
				{
					Key:   "TotalAssets",
					Label: "Total Assets",
					Values: []domain.StatementValue{
						{Value: 352_583_000_000, Present: true},
						{Value: 337_411_000_000, Present: true},
					},
				},
				{
					Key:   "TotalDebt",
					Label: "Total Debt",
					Values: []domain.StatementValue{
						{Value: 123_930_000_000, Present: true},
						{Value: 111_088_000_000, Present: true},
					},
				},
				{
					Key:   "CashAndCashEquivalents",
					Label: "Cash And Cash Equivalents",
					Values: []domain.StatementValue{
						{Value: 29_943_000_000, Present: true},
						{Value: 29_965_000_000, Present: true},
					},
				},
			}
		case domain.StatementKindCashFlow:
			stmt.Rows = []domain.StatementRow{
				{
					Key:   "OperatingCashFlow",
					Label: "Operating Cash Flow",
					Values: []domain.StatementValue{
						{Value: 118_254_000_000, Present: true},
						{Value: 110_543_000_000, Present: true},
					},
				},
				{
					Key:   "FreeCashFlow",
					Label: "Free Cash Flow",
					Values: []domain.StatementValue{
						{Value: 99_584_000_000, Present: true},
						{Value: 98_995_000_000, Present: true},
					},
				},
			}
		}
		m.cacheStatement(stmt)
		if req.kind == m.statementKind && req.frequency == m.statementFreq {
			m.statement = stmt
		}
	}
}
