package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"blackdesk/internal/agents"
	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

type AIConnector struct {
	ID    string
	Label string
}

type Services struct {
	registry    *providers.Registry
	agents      *agents.Registry
	configStore *storage.ConfigStore
	filings     FilingsProvider
}

var errServiceUnavailable = errors.New("service unavailable")

func NewServices(registry *providers.Registry, agentsRegistry *agents.Registry, configStore *storage.ConfigStore) *Services {
	return NewServicesWithFilings(registry, agentsRegistry, configStore, nil)
}

func NewServicesWithFilings(registry *providers.Registry, agentsRegistry *agents.Registry, configStore *storage.ConfigStore, filingsProvider FilingsProvider) *Services {
	return &Services{
		registry:    registry,
		agents:      agentsRegistry,
		configStore: configStore,
		filings:     filingsProvider,
	}
}

func (s *Services) ActiveProviderName() string {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return ""
	}
	return strings.TrimSpace(s.registry.Active().Name())
}

func (s *Services) GetQuote(ctx context.Context, symbol string) (domain.QuoteSnapshot, error) {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return domain.QuoteSnapshot{}, errServiceUnavailable
	}
	return s.registry.Active().GetQuote(ctx, symbol)
}

func (s *Services) GetQuotes(ctx context.Context, symbols []string) ([]domain.QuoteSnapshot, error) {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return nil, errServiceUnavailable
	}
	return s.registry.Active().GetQuotes(ctx, symbols)
}

func (s *Services) GetHistory(ctx context.Context, symbol, timeRange, interval string) (domain.PriceSeries, error) {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return domain.PriceSeries{}, errServiceUnavailable
	}
	return s.registry.Active().GetHistory(ctx, symbol, timeRange, interval)
}

func (s *Services) GetNews(ctx context.Context, symbol string) ([]domain.NewsItem, error) {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return nil, errServiceUnavailable
	}
	return s.registry.Active().GetNews(ctx, symbol)
}

func (s *Services) GetMarketNews(ctx context.Context) ([]domain.NewsItem, []domain.MarketNewsSource, error) {
	if s == nil || s.registry == nil {
		return nil, nil, nil
	}
	provider, ok := s.registry.MarketNews()
	if !ok {
		return nil, nil, nil
	}
	items, err := provider.GetMarketNews(ctx)
	var sources []domain.MarketNewsSource
	if catalog, ok := provider.(providers.MarketNewsCatalogProvider); ok {
		sources = catalog.MarketNewsSources()
	}
	return items, sources, err
}

func (s *Services) GetFundamentals(ctx context.Context, symbol string) (domain.FundamentalsSnapshot, error) {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return domain.FundamentalsSnapshot{}, errServiceUnavailable
	}
	return s.registry.Active().GetFundamentals(ctx, symbol)
}

func (s *Services) SearchSymbols(ctx context.Context, query string) ([]domain.SymbolRef, error) {
	if s == nil || s.registry == nil || s.registry.Active() == nil {
		return nil, errServiceUnavailable
	}
	return s.registry.Active().SearchSymbols(ctx, query)
}

func (s *Services) HasStatements() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.Statements()
	return ok
}

func (s *Services) GetStatement(ctx context.Context, symbol string, kind domain.StatementKind, frequency domain.StatementFrequency) (domain.FinancialStatement, error) {
	if s == nil || s.registry == nil {
		return domain.FinancialStatement{}, errServiceUnavailable
	}
	provider, _ := s.registry.Statements()
	if provider == nil {
		return domain.FinancialStatement{}, errServiceUnavailable
	}
	return provider.GetStatement(ctx, symbol, kind, frequency)
}

func (s *Services) HasInsiders() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.Insiders()
	return ok
}

func (s *Services) GetInsiders(ctx context.Context, symbol string) (domain.InsiderSnapshot, error) {
	if s == nil || s.registry == nil {
		return domain.InsiderSnapshot{}, errServiceUnavailable
	}
	provider, _ := s.registry.Insiders()
	if provider == nil {
		return domain.InsiderSnapshot{}, errServiceUnavailable
	}
	return provider.GetInsiders(ctx, symbol)
}

func (s *Services) HasOwners() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.Owners()
	return ok
}

func (s *Services) GetOwners(ctx context.Context, symbol string) (domain.OwnershipSnapshot, error) {
	if s == nil || s.registry == nil {
		return domain.OwnershipSnapshot{}, errServiceUnavailable
	}
	provider, _ := s.registry.Owners()
	if provider == nil {
		return domain.OwnershipSnapshot{}, errServiceUnavailable
	}
	return provider.GetOwners(ctx, symbol)
}

func (s *Services) HasAnalystRecommendations() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.AnalystRecommendations()
	return ok
}

func (s *Services) GetAnalystRecommendations(ctx context.Context, symbol string) (domain.AnalystRecommendationsSnapshot, error) {
	if s == nil || s.registry == nil {
		return domain.AnalystRecommendationsSnapshot{}, errServiceUnavailable
	}
	provider, _ := s.registry.AnalystRecommendations()
	if provider == nil {
		return domain.AnalystRecommendationsSnapshot{}, errServiceUnavailable
	}
	return provider.GetAnalystRecommendations(ctx, symbol)
}

func (s *Services) HasEarnings() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.Earnings()
	return ok
}

func (s *Services) GetEarnings(ctx context.Context, symbol string) (domain.EarningsSnapshot, error) {
	if s == nil || s.registry == nil {
		return domain.EarningsSnapshot{}, errServiceUnavailable
	}
	provider, _ := s.registry.Earnings()
	if provider == nil {
		return domain.EarningsSnapshot{}, errServiceUnavailable
	}
	return provider.GetEarnings(ctx, symbol)
}

func (s *Services) HasEconomicCalendar() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.EconomicCalendar()
	return ok
}

func (s *Services) GetEconomicCalendar(ctx context.Context, start, end time.Time) (domain.EconomicCalendarSnapshot, error) {
	if s == nil || s.registry == nil {
		return domain.EconomicCalendarSnapshot{}, errServiceUnavailable
	}
	provider, _ := s.registry.EconomicCalendar()
	if provider == nil {
		return domain.EconomicCalendarSnapshot{}, errServiceUnavailable
	}
	return provider.GetEconomicCalendar(ctx, start, end)
}

func (s *Services) HasScreeners() bool {
	if s == nil || s.registry == nil {
		return false
	}
	_, ok := s.registry.Screeners()
	return ok
}

func (s *Services) Screeners() []domain.ScreenerDefinition {
	if s == nil || s.registry == nil {
		return nil
	}
	provider, ok := s.registry.Screeners()
	if !ok {
		return nil
	}
	return provider.Screeners()
}

func (s *Services) GetScreener(ctx context.Context, id string, count int) (domain.ScreenerResult, error) {
	if s == nil || s.registry == nil {
		return domain.ScreenerResult{}, errServiceUnavailable
	}
	provider, _ := s.registry.Screeners()
	if provider == nil {
		return domain.ScreenerResult{}, errServiceUnavailable
	}
	return provider.GetScreener(ctx, id, count)
}

func (s *Services) HasFilings() bool {
	return s != nil && s.filings != nil
}

func (s *Services) GetFilings(ctx context.Context, symbol string) (domain.FilingsSnapshot, error) {
	if s == nil || s.filings == nil {
		return domain.FilingsSnapshot{}, errServiceUnavailable
	}
	return s.filings.GetFilings(ctx, symbol)
}

func (s *Services) GetFilingDocument(ctx context.Context, item domain.FilingItem) (domain.FilingDocument, error) {
	if s == nil || s.filings == nil {
		return domain.FilingDocument{}, errServiceUnavailable
	}
	return s.filings.GetFilingDocument(ctx, item)
}

func (s *Services) SaveConfig(cfg storage.Config) error {
	if s == nil || s.configStore == nil {
		return nil
	}
	return s.configStore.Save(cfg)
}

func (s *Services) ListAIConnectors() []AIConnector {
	if s == nil || s.agents == nil {
		return nil
	}
	items := s.agents.List()
	out := make([]AIConnector, 0, len(items))
	for _, item := range items {
		out = append(out, AIConnector{ID: item.ID, Label: item.Label})
	}
	return out
}

func (s *Services) LookupAIConnector(id string) (AIConnector, bool) {
	if s == nil || s.agents == nil {
		return AIConnector{}, false
	}
	item, ok := s.agents.Lookup(id)
	if !ok {
		return AIConnector{}, false
	}
	return AIConnector{ID: item.ID, Label: item.Label}, true
}

func (s *Services) RunAI(ctx context.Context, connectorID string, req agents.Request) (agents.Response, error) {
	if s == nil || s.agents == nil {
		return agents.Response{}, errServiceUnavailable
	}
	return s.agents.Run(ctx, connectorID, req)
}

func (s *Services) AIModels(ctx context.Context, connectorID string) ([]string, error) {
	if s == nil || s.agents == nil {
		return nil, errServiceUnavailable
	}
	return s.agents.Models(ctx, connectorID)
}
