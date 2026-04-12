package providers

import (
	"context"

	"blackdesk/internal/domain"
)

type QuoteProvider interface {
	GetQuote(context.Context, string) (domain.QuoteSnapshot, error)
	GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error)
}

type HistoryProvider interface {
	GetHistory(context.Context, string, string, string) (domain.PriceSeries, error)
}

type NewsProvider interface {
	GetNews(context.Context, string) ([]domain.NewsItem, error)
}

type MarketNewsProvider interface {
	GetMarketNews(context.Context) ([]domain.NewsItem, error)
}

type MarketNewsCatalogProvider interface {
	MarketNewsSources() []domain.MarketNewsSource
}

type FundamentalsProvider interface {
	GetFundamentals(context.Context, string) (domain.FundamentalsSnapshot, error)
}

type StatementsProvider interface {
	GetStatement(context.Context, string, domain.StatementKind, domain.StatementFrequency) (domain.FinancialStatement, error)
}

type InsidersProvider interface {
	GetInsiders(context.Context, string) (domain.InsiderSnapshot, error)
}

type EarningsProvider interface {
	GetEarnings(context.Context, string) (domain.EarningsSnapshot, error)
}

type SymbolSearchProvider interface {
	SearchSymbols(context.Context, string) ([]domain.SymbolRef, error)
}

type ScreenerProvider interface {
	Screeners() []domain.ScreenerDefinition
	GetScreener(context.Context, string, int) (domain.ScreenerResult, error)
}

type Provider interface {
	Name() string
	Capabilities() domain.ProviderCapabilities
	QuoteProvider
	HistoryProvider
	NewsProvider
	FundamentalsProvider
	SymbolSearchProvider
}

type Registry struct {
	active Provider
}

func NewRegistry(active Provider) *Registry {
	return &Registry{active: active}
}

func (r *Registry) Active() Provider {
	return r.active
}

func (r *Registry) Statements() (StatementsProvider, bool) {
	if r == nil || r.active == nil {
		return nil, false
	}
	p, ok := r.active.(StatementsProvider)
	return p, ok
}

func (r *Registry) Insiders() (InsidersProvider, bool) {
	if r == nil || r.active == nil {
		return nil, false
	}
	p, ok := r.active.(InsidersProvider)
	return p, ok
}

func (r *Registry) Earnings() (EarningsProvider, bool) {
	if r == nil || r.active == nil {
		return nil, false
	}
	p, ok := r.active.(EarningsProvider)
	return p, ok
}

func (r *Registry) MarketNews() (MarketNewsProvider, bool) {
	if r == nil || r.active == nil {
		return nil, false
	}
	p, ok := r.active.(MarketNewsProvider)
	return p, ok
}

func (r *Registry) Screeners() (ScreenerProvider, bool) {
	if r == nil || r.active == nil {
		return nil, false
	}
	p, ok := r.active.(ScreenerProvider)
	return p, ok
}
