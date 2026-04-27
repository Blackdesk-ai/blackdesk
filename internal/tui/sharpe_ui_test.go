package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestCommandPaletteSharpeOpensQuoteSharpeMode(t *testing.T) {
	provider := &countingHistoryProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{{Kind: commandPaletteItemFunction, FunctionID: "sharpe", Title: "Risk Adjusted"}}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterSharpe {
		t.Fatalf("expected Quote sharpe mode, got tab=%d mode=%d", m.tabIdx, m.quoteCenterMode)
	}
	if cmd == nil {
		t.Fatal("expected sharpe load command")
	}
	if msg, ok := cmd().(sharpeHistoryLoadedMsg); !ok {
		t.Fatalf("expected sharpeHistoryLoadedMsg, got %T", msg)
	}
	if len(provider.historyCalls) != 1 || provider.historyCalls[0] != "AAPL|10y|1d" {
		t.Fatalf("expected 10y sharpe history request, got %#v", provider.historyCalls)
	}
	if ranges[m.sharpeRangeIdx].Range != "5y" {
		t.Fatalf("expected sharpe range to switch to 5y, got %q", ranges[m.sharpeRangeIdx].Range)
	}
}

func TestSharpeHistoryFallsBackFrom10YTo7Y(t *testing.T) {
	provider := &fallbackSharpeHistoryProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})

	series, err := model.loadSharpeHistory("AAPL")
	if err != nil {
		t.Fatalf("expected fallback history load to succeed, got %v", err)
	}
	if series.Range != "7y" {
		t.Fatalf("expected fallback series range 7y, got %q", series.Range)
	}
	if got := provider.calls; len(got) != 2 || got[0] != "AAPL|10y|1d" || got[1] != "AAPL|7y|1d" {
		t.Fatalf("expected 10y then 7y fallback calls, got %#v", got)
	}
	model.sharpeCache["AAPL"] = series
	if model.needsSharpeHistory("AAPL") {
		t.Fatal("expected 7y sharpe cache to be considered valid")
	}
}

func TestQuoteSharpeViewRendersFullscreenChartAndPreview(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterSharpe
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.quote.Symbol = "AAPL"
	model.quote.ShortName = "Apple Inc."
	model.sharpeCache["AAPL"] = sampleSharpeHistorySeries("AAPL")

	view := model.View()
	if !strings.Contains(view, "RISK ADJUSTED") || !strings.Contains(view, "Latest") {
		t.Fatal("expected sharpe board and preview section")
	}
	if !strings.Contains(view, "252d") || !strings.Contains(view, "63d") {
		t.Fatal("expected sharpe mode to show both 252d and 63d sharpe series")
	}
	if !strings.Contains(view, "3M Avg Return") || !strings.Contains(view, "3M Median Return") || !strings.Contains(view, "3M Price Higher") {
		t.Fatal("expected sharpe preview to show forward 3M return stats")
	}
	if !strings.Contains(view, "TIMEFRAMES") || !strings.Contains(view, "←/→") {
		t.Fatal("expected sharpe board to render chart-style timeframe controls")
	}
	if strings.Contains(view, "12M Sharpe path") || strings.Contains(view, "Formula") || strings.Contains(view, "ROC/HV") {
		t.Fatal("expected sharpe preview to omit introductory text")
	}
	if strings.Contains(view, "MARKET HEAT") || strings.Contains(view, "ANALYSTS") || strings.Contains(view, "AI INSIGHT") {
		t.Fatal("expected sharpe mode to replace default right sidebar content")
	}
}

func TestCommandPaletteStatisticsOpensStatisticsMode(t *testing.T) {
	provider := &countingHistoryProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{{Kind: commandPaletteItemFunction, FunctionID: "statistics", Title: "Statistics"}}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterStatistics {
		t.Fatalf("expected Quote statistics mode, got tab=%d mode=%d", m.tabIdx, m.quoteCenterMode)
	}
	if cmd == nil {
		t.Fatal("expected statistics history load command")
	}
	if msg, ok := cmd().(sharpeHistoryLoadedMsg); !ok {
		t.Fatalf("expected sharpeHistoryLoadedMsg, got %T", msg)
	}
	if len(provider.historyCalls) != 1 || provider.historyCalls[0] != "AAPL|5y|1d" {
		t.Fatalf("expected 5y statistics history request, got %#v", provider.historyCalls)
	}
}

func TestQuoteStatisticsViewRendersForwardReturnStats(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterStatistics
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.quote.Symbol = "AAPL"
	model.quote.ShortName = "Apple Inc."
	model.sharpeCache["AAPL"] = sampleSharpeHistorySeries("AAPL")

	view := model.View()
	for _, want := range []string{"STATISTICS", "FORWARD RETURNS (vs ROC/HV)", "5Y", "Max", "Date", "Signal", "Avg", "Median", "Win%", "12M > 0", "12M"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected statistics view to contain %q", want)
		}
	}
	for _, unwanted := range []string{"Higher", "Best", "Worst"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("expected statistics view to omit %q", unwanted)
		}
	}
	if strings.Contains(view, "MARKET HEAT") || strings.Contains(view, "ANALYSTS") {
		t.Fatal("expected statistics mode to replace default right sidebar content")
	}
}

func TestStatisticsRangeNavigationLoadsMaxHistory(t *testing.T) {
	provider := &countingHistoryProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterStatistics
	model.statisticsRangeIdx = 0

	updated, cmd, handled := model.handleGlobalNavigationKey("right")
	if !handled {
		t.Fatal("expected statistics range navigation to handle right arrow")
	}
	if updated.statisticsRangeIdx != 1 {
		t.Fatalf("expected Max range index, got %d", updated.statisticsRangeIdx)
	}
	if cmd == nil {
		t.Fatal("expected max statistics history load command")
	}
	if msg, ok := cmd().(sharpeHistoryLoadedMsg); !ok {
		t.Fatalf("expected sharpeHistoryLoadedMsg, got %T", msg)
	}
	if len(provider.historyCalls) != 1 || provider.historyCalls[0] != "AAPL|10y|1d" {
		t.Fatalf("expected 10y max statistics history request, got %#v", provider.historyCalls)
	}
}

func sampleSharpeHistorySeries(symbol string) domain.PriceSeries {
	candles := make([]domain.Candle, 0, 900)
	base := time.Now().AddDate(-4, 0, 0)
	price := 100.0
	for i := 0; i < 900; i++ {
		price += 0.08
		if i%17 == 0 {
			price -= 1.2
		}
		if i%43 == 0 {
			price += 2.1
		}
		candles = append(candles, domain.Candle{
			Time:   base.AddDate(0, 0, i),
			Open:   price - 0.5,
			High:   price + 0.7,
			Low:    price - 0.9,
			Close:  price,
			Volume: 1_000_000,
		})
	}
	return domain.PriceSeries{Symbol: symbol, Range: "5y", Interval: "1d", Candles: candles}
}

type fallbackSharpeHistoryProvider struct {
	calls []string
}

func (p *fallbackSharpeHistoryProvider) Name() string { return "test" }

func (p *fallbackSharpeHistoryProvider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{}
}

func (p *fallbackSharpeHistoryProvider) GetQuote(context.Context, string) (domain.QuoteSnapshot, error) {
	return domain.QuoteSnapshot{}, nil
}

func (p *fallbackSharpeHistoryProvider) GetQuotes(context.Context, []string) ([]domain.QuoteSnapshot, error) {
	return nil, nil
}

func (p *fallbackSharpeHistoryProvider) GetHistory(_ context.Context, symbol, rangeKey, interval string) (domain.PriceSeries, error) {
	p.calls = append(p.calls, symbol+"|"+rangeKey+"|"+interval)
	if rangeKey == "10y" {
		return domain.PriceSeries{}, errors.New("chart candles empty")
	}
	series := sampleSharpeHistorySeries(symbol)
	series.Range = rangeKey
	return series, nil
}

func (p *fallbackSharpeHistoryProvider) GetNews(context.Context, string) ([]domain.NewsItem, error) {
	return nil, nil
}

func (p *fallbackSharpeHistoryProvider) GetFundamentals(context.Context, string) (domain.FundamentalsSnapshot, error) {
	return domain.FundamentalsSnapshot{}, nil
}

func (p *fallbackSharpeHistoryProvider) GetAnalystRecommendations(context.Context, string) (domain.AnalystRecommendationsSnapshot, error) {
	return domain.AnalystRecommendationsSnapshot{}, nil
}

func (p *fallbackSharpeHistoryProvider) GetOwners(context.Context, string) (domain.OwnershipSnapshot, error) {
	return domain.OwnershipSnapshot{}, nil
}

func (p *fallbackSharpeHistoryProvider) GetEarnings(context.Context, string) (domain.EarningsSnapshot, error) {
	return domain.EarningsSnapshot{}, nil
}

func (p *fallbackSharpeHistoryProvider) GetEconomicCalendar(context.Context, time.Time, time.Time) (domain.EconomicCalendarSnapshot, error) {
	return domain.EconomicCalendarSnapshot{}, nil
}

func (p *fallbackSharpeHistoryProvider) SearchSymbols(context.Context, string) ([]domain.SymbolRef, error) {
	return nil, nil
}
