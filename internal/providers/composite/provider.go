package composite

import (
	"context"
	"fmt"
	"time"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
)

type Provider struct {
	base       providers.Provider
	marketNews providers.MarketNewsProvider
}

func New(base providers.Provider, marketNews providers.MarketNewsProvider) *Provider {
	return &Provider{
		base:       base,
		marketNews: marketNews,
	}
}

func (p *Provider) Name() string {
	if p == nil || p.base == nil {
		return ""
	}
	return p.base.Name()
}

func (p *Provider) Capabilities() domain.ProviderCapabilities {
	if p == nil || p.base == nil {
		return domain.ProviderCapabilities{}
	}
	caps := p.base.Capabilities()
	caps.MarketNews = p.marketNews != nil
	return caps
}

func (p *Provider) GetQuote(ctx context.Context, symbol string) (domain.QuoteSnapshot, error) {
	if p == nil || p.base == nil {
		return domain.QuoteSnapshot{}, fmt.Errorf("base provider unavailable")
	}
	return p.base.GetQuote(ctx, symbol)
}

func (p *Provider) GetQuotes(ctx context.Context, symbols []string) ([]domain.QuoteSnapshot, error) {
	if p == nil || p.base == nil {
		return nil, fmt.Errorf("base provider unavailable")
	}
	return p.base.GetQuotes(ctx, symbols)
}

func (p *Provider) GetHistory(ctx context.Context, symbol, rangeKey, interval string) (domain.PriceSeries, error) {
	if p == nil || p.base == nil {
		return domain.PriceSeries{}, fmt.Errorf("base provider unavailable")
	}
	return p.base.GetHistory(ctx, symbol, rangeKey, interval)
}

func (p *Provider) GetNews(ctx context.Context, symbol string) ([]domain.NewsItem, error) {
	if p == nil || p.base == nil {
		return nil, fmt.Errorf("base provider unavailable")
	}
	return p.base.GetNews(ctx, symbol)
}

func (p *Provider) GetFundamentals(ctx context.Context, symbol string) (domain.FundamentalsSnapshot, error) {
	if p == nil || p.base == nil {
		return domain.FundamentalsSnapshot{}, fmt.Errorf("base provider unavailable")
	}
	return p.base.GetFundamentals(ctx, symbol)
}

func (p *Provider) SearchSymbols(ctx context.Context, query string) ([]domain.SymbolRef, error) {
	if p == nil || p.base == nil {
		return nil, fmt.Errorf("base provider unavailable")
	}
	return p.base.SearchSymbols(ctx, query)
}

func (p *Provider) GetMarketNews(ctx context.Context) ([]domain.NewsItem, error) {
	if p == nil || p.marketNews == nil {
		return nil, fmt.Errorf("market news provider unavailable")
	}
	return p.marketNews.GetMarketNews(ctx)
}

func (p *Provider) MarketNewsSources() []domain.MarketNewsSource {
	if p == nil || p.marketNews == nil {
		return nil
	}
	catalog, ok := p.marketNews.(providers.MarketNewsCatalogProvider)
	if !ok {
		return nil
	}
	return catalog.MarketNewsSources()
}

func (p *Provider) GetStatement(ctx context.Context, symbol string, kind domain.StatementKind, frequency domain.StatementFrequency) (domain.FinancialStatement, error) {
	statementsProvider, ok := p.base.(providers.StatementsProvider)
	if !ok {
		return domain.FinancialStatement{}, fmt.Errorf("statements provider unavailable")
	}
	return statementsProvider.GetStatement(ctx, symbol, kind, frequency)
}

func (p *Provider) GetInsiders(ctx context.Context, symbol string) (domain.InsiderSnapshot, error) {
	insidersProvider, ok := p.base.(providers.InsidersProvider)
	if !ok {
		return domain.InsiderSnapshot{}, fmt.Errorf("insiders provider unavailable")
	}
	return insidersProvider.GetInsiders(ctx, symbol)
}

func (p *Provider) GetEarnings(ctx context.Context, symbol string) (domain.EarningsSnapshot, error) {
	earningsProvider, ok := p.base.(providers.EarningsProvider)
	if !ok {
		return domain.EarningsSnapshot{}, fmt.Errorf("earnings provider unavailable")
	}
	return earningsProvider.GetEarnings(ctx, symbol)
}

func (p *Provider) GetEconomicCalendar(ctx context.Context, start, end time.Time) (domain.EconomicCalendarSnapshot, error) {
	calendarProvider, ok := p.base.(providers.EconomicCalendarProvider)
	if !ok {
		return domain.EconomicCalendarSnapshot{}, fmt.Errorf("economic calendar provider unavailable")
	}
	return calendarProvider.GetEconomicCalendar(ctx, start, end)
}

func (p *Provider) Screeners() []domain.ScreenerDefinition {
	screenerProvider, ok := p.base.(providers.ScreenerProvider)
	if !ok {
		return nil
	}
	return screenerProvider.Screeners()
}

func (p *Provider) GetScreener(ctx context.Context, id string, count int) (domain.ScreenerResult, error) {
	screenerProvider, ok := p.base.(providers.ScreenerProvider)
	if !ok {
		return domain.ScreenerResult{}, fmt.Errorf("screener provider unavailable")
	}
	return screenerProvider.GetScreener(ctx, id, count)
}
