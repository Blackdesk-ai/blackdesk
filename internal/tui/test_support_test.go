package tui

import (
	"context"
	"strings"
	"time"

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
