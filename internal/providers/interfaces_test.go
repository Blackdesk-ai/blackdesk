package providers

import (
	"context"
	"testing"
	"time"

	"blackdesk/internal/domain"
)

type baseProvider struct{}

func (baseProvider) Name() string { return "base" }

func (baseProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{}
}

func (baseProvider) GetQuote(context.Context, string) (domain.QuoteSnapshot, error) {
	return domain.QuoteSnapshot{}, nil
}

func (baseProvider) GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error) {
	return nil, nil
}

func (baseProvider) GetHistory(context.Context, string, string, string) (domain.PriceSeries, error) {
	return domain.PriceSeries{}, nil
}

func (baseProvider) GetNews(context.Context, string) ([]domain.NewsItem, error) {
	return nil, nil
}

func (baseProvider) GetFundamentals(context.Context, string) (domain.FundamentalsSnapshot, error) {
	return domain.FundamentalsSnapshot{}, nil
}

func (baseProvider) SearchSymbols(context.Context, string) ([]domain.SymbolRef, error) {
	return nil, nil
}

type statementsCapableProvider struct {
	baseProvider
}

type marketNewsCapableProvider struct {
	baseProvider
}

type insidersCapableProvider struct {
	baseProvider
}

type ownersCapableProvider struct {
	baseProvider
}

type analystCapableProvider struct {
	baseProvider
}

type screenerCapableProvider struct {
	baseProvider
}

type economicCalendarCapableProvider struct {
	baseProvider
}

func (statementsCapableProvider) GetStatement(context.Context, string, domain.StatementKind, domain.StatementFrequency) (domain.FinancialStatement, error) {
	return domain.FinancialStatement{}, nil
}

func (insidersCapableProvider) GetInsiders(context.Context, string) (domain.InsiderSnapshot, error) {
	return domain.InsiderSnapshot{}, nil
}

func (ownersCapableProvider) GetOwners(context.Context, string) (domain.OwnershipSnapshot, error) {
	return domain.OwnershipSnapshot{}, nil
}

func (analystCapableProvider) GetAnalystRecommendations(context.Context, string) (domain.AnalystRecommendationsSnapshot, error) {
	return domain.AnalystRecommendationsSnapshot{}, nil
}

func (screenerCapableProvider) Screeners() []domain.ScreenerDefinition {
	return []domain.ScreenerDefinition{{ID: "most_actives"}}
}

func (screenerCapableProvider) GetScreener(context.Context, string, int) (domain.ScreenerResult, error) {
	return domain.ScreenerResult{}, nil
}

func (marketNewsCapableProvider) GetMarketNews(context.Context) ([]domain.NewsItem, error) {
	return nil, nil
}

func (economicCalendarCapableProvider) GetEconomicCalendar(context.Context, time.Time, time.Time) (domain.EconomicCalendarSnapshot, error) {
	return domain.EconomicCalendarSnapshot{}, nil
}

func TestRegistryStatementsReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(statementsCapableProvider{})

	p, ok := registry.Statements()
	if !ok || p == nil {
		t.Fatal("expected statements provider to be exposed")
	}
}

func TestRegistryStatementsHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.Statements()
	if ok || p != nil {
		t.Fatal("expected missing statements provider to return false")
	}
}

func TestRegistryStatementsHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.Statements()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}

func TestRegistryMarketNewsReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(marketNewsCapableProvider{})

	p, ok := registry.MarketNews()
	if !ok || p == nil {
		t.Fatal("expected market news provider to be exposed")
	}
}

func TestRegistryMarketNewsHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.MarketNews()
	if ok || p != nil {
		t.Fatal("expected missing market news provider to return false")
	}
}

func TestRegistryMarketNewsHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.MarketNews()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}

func TestRegistryInsidersReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(insidersCapableProvider{})

	p, ok := registry.Insiders()
	if !ok || p == nil {
		t.Fatal("expected insiders provider to be exposed")
	}
}

func TestRegistryInsidersHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.Insiders()
	if ok || p != nil {
		t.Fatal("expected missing insiders provider to return false")
	}
}

func TestRegistryInsidersHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.Insiders()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}

func TestRegistryOwnersReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(ownersCapableProvider{})

	p, ok := registry.Owners()
	if !ok || p == nil {
		t.Fatal("expected owners provider to be exposed")
	}
}

func TestRegistryOwnersHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.Owners()
	if ok || p != nil {
		t.Fatal("expected missing owners provider to return false")
	}
}

func TestRegistryOwnersHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.Owners()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}

func TestRegistryAnalystRecommendationsReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(analystCapableProvider{})

	p, ok := registry.AnalystRecommendations()
	if !ok || p == nil {
		t.Fatal("expected analyst recommendations provider to be exposed")
	}
}

func TestRegistryAnalystRecommendationsHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.AnalystRecommendations()
	if ok || p != nil {
		t.Fatal("expected missing analyst recommendations provider to return false")
	}
}

func TestRegistryAnalystRecommendationsHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.AnalystRecommendations()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}

func TestRegistryEconomicCalendarReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(economicCalendarCapableProvider{})

	p, ok := registry.EconomicCalendar()
	if !ok || p == nil {
		t.Fatal("expected economic calendar provider to be exposed")
	}
}

func TestRegistryEconomicCalendarHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.EconomicCalendar()
	if ok || p != nil {
		t.Fatal("expected missing economic calendar provider to return false")
	}
}

func TestRegistryEconomicCalendarHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.EconomicCalendar()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}

func TestRegistryScreenersReturnsOptionalProvider(t *testing.T) {
	registry := NewRegistry(screenerCapableProvider{})

	p, ok := registry.Screeners()
	if !ok || p == nil {
		t.Fatal("expected screener provider to be exposed")
	}
}

func TestRegistryScreenersHandlesMissingCapability(t *testing.T) {
	registry := NewRegistry(baseProvider{})

	p, ok := registry.Screeners()
	if ok || p != nil {
		t.Fatal("expected missing screener provider to return false")
	}
}

func TestRegistryScreenersHandlesNilRegistry(t *testing.T) {
	var registry *Registry

	p, ok := registry.Screeners()
	if ok || p != nil {
		t.Fatal("expected nil registry to return false")
	}
}
