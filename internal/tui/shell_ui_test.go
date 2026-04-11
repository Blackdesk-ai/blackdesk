package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestHeaderRendersBrandMark(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 30
	model.quote = domain.QuoteSnapshot{Symbol: "AAPL", Price: 210}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "|- BLACKDESK") {
		t.Fatal("expected header to render the compact brand mark")
	}
}

func TestViewShowsStatusMetaOnStatusLine(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 120
	model.height = 40
	model.status = "Loaded quote for BRK-B"

	view := model.View()
	if !strings.Contains(view, "Loaded quote for BRK-B") {
		t.Fatal("expected left status text")
	}
	if !strings.Contains(view, "Source: test") {
		t.Fatal("expected right-aligned status metadata")
	}
}

func TestMarketStatusHidesQuoteOnlyKeys(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 120
	model.height = 40
	model.tabIdx = tabMarkets

	status := ansi.Strip(model.statusText())
	if !strings.Contains(status, "Tab tabs") {
		t.Fatal("expected market status to keep tab navigation")
	}
	if strings.Contains(status, "↑/↓ symbols") {
		t.Fatal("expected market status to hide symbol navigation")
	}
	if strings.Contains(status, "←/→ timeframe") {
		t.Fatal("expected market status to hide timeframe navigation")
	}
	if strings.Contains(status, "fundamentals") || strings.Contains(status, "technicals") {
		t.Fatal("expected market status to hide quote center toggles")
	}
	if strings.Contains(status, "news") || strings.Contains(status, "profile") || strings.Contains(status, "open news") {
		t.Fatal("expected market status to hide quote research navigation")
	}
}

func TestViewRendersTabSpecificContent(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.quote = domain.QuoteSnapshot{
		Symbol:            "AAPL",
		ShortName:         "Apple",
		Currency:          "USD",
		Price:             210.12,
		Change:            3.15,
		ChangePercent:     1.52,
		Open:              208.10,
		DayLow:            207.55,
		DayHigh:           211.33,
		Volume:            91_000_000,
		AverageVolume:     72_000_000,
		MarketCap:         3_200_000_000_000,
		Exchange:          "NMS",
		MarketState:       domain.MarketStateRegular,
		PreviousClose:     206.97,
		RegularMarketTime: time.Now(),
	}
	model.series = domain.PriceSeries{
		Symbol:   "AAPL",
		Range:    "3mo",
		Interval: "1d",
		Candles: []domain.Candle{
			{Time: time.Now().AddDate(0, 0, -5), Open: 190, High: 193, Low: 189, Close: 192},
			{Time: time.Now().AddDate(0, 0, -4), Open: 192, High: 196, Low: 191, Close: 195},
			{Time: time.Now().AddDate(0, 0, -3), Open: 195, High: 199, Low: 194, Close: 198},
			{Time: time.Now().AddDate(0, 0, -2), Open: 198, High: 204, Low: 197, Close: 202},
			{Time: time.Now().AddDate(0, 0, -1), Open: 202, High: 211, Low: 201, Close: 210},
		},
	}
	model.news = []domain.NewsItem{
		{Title: "Apple expands AI push across devices", Publisher: "Reuters", Time: time.Now()},
	}
	model.fundamentals = domain.FundamentalsSnapshot{
		Symbol:                  "AAPL",
		Sector:                  "Technology",
		Industry:                "Consumer Electronics",
		Description:             "Apple designs consumer devices, software, and services across a tightly integrated ecosystem.",
		MarketCap:               3_200_000_000_000,
		EnterpriseValue:         3_280_000_000_000,
		TrailingPE:              31.2,
		ForwardPE:               28.4,
		PEGRatio:                2.14,
		PriceToSales:            8.10,
		EnterpriseToRevenue:     8.30,
		EnterpriseToEBITDA:      24.5,
		BookValue:               4.25,
		TrailingEPS:             4.90,
		EPS:                     4.90,
		RevenuePerShare:         25.1,
		DividendYield:           0.0048,
		FiftyTwoWeekLow:         164.08,
		FiftyTwoWeekHigh:        212.40,
		AverageVolume:           72_000_000,
		Beta:                    1.18,
		PriceToBook:             45.1,
		Revenue:                 391_000_000_000,
		GrossProfits:            180_000_000_000,
		EBITDA:                  134_000_000_000,
		OperatingCashflow:       122_000_000_000,
		FreeCashflow:            99_000_000_000,
		TotalCash:               67_000_000_000,
		TotalDebt:               110_000_000_000,
		GrossMargins:            0.462,
		ProfitMargins:           0.262,
		OperatingMargins:        0.311,
		ReturnOnAssets:          0.284,
		ReturnOnEquity:          1.55,
		ReturnOnInvestedCapital: 0.421,
		RevenueGrowth:           0.061,
		EarningsGrowth:          0.104,
		CurrentRatio:            0.99,
		QuickRatio:              0.83,
		DebtToEquity:            167.0,
		RecommendationMean:      1.9,
		RecommendationKey:       "buy",
		AnalystOpinions:         39,
		TargetLowPrice:          185,
		TargetMeanPrice:         225,
		TargetHighPrice:         250,
	}

	model.tabIdx = tabQuote
	quoteView := model.View()
	if !strings.Contains(quoteView, "TIMEFRAMES") {
		t.Fatal("expected quote tab controls")
	}
	if !strings.Contains(quoteView, "MARKET HEAT") {
		t.Fatal("expected quote right panel")
	}
	if !strings.Contains(quoteView, "PEG") {
		t.Fatal("expected PEG ratio in quote fundamentals sidebar")
	}
	if strings.Contains(quoteView, "Mean") {
		t.Fatal("expected Mean to be removed from quote analysts sidebar")
	}
	if strings.Contains(quoteView, "Opinions") {
		t.Fatal("expected Opinions to be removed from quote analysts sidebar")
	}
	if strings.Contains(quoteView, "Industry") {
		t.Fatal("expected Industry to be removed from quote fundamentals sidebar")
	}
	if strings.Contains(quoteView, "Margins G/P") {
		t.Fatal("expected Margins to be removed from quote fundamentals sidebar")
	}
	if !strings.Contains(quoteView, "NEWS") {
		t.Fatal("expected quote bottom news panel")
	}
	if !strings.Contains(ansi.Strip(quoteView), "NEWS (n)") {
		t.Fatal("expected quote bottom news panel key hint")
	}
	if !strings.Contains(quoteView, "PROFILE") {
		t.Fatal("expected quote bottom profile panel")
	}
	if !strings.Contains(ansi.Strip(quoteView), "PROFILE (p)") {
		t.Fatal("expected quote bottom profile panel key hint")
	}
	if !strings.Contains(ansi.Strip(quoteView), "Technology") {
		t.Fatal("expected sector value in profile header")
	}
	if strings.Contains(ansi.Strip(quoteView), "Sector Technology") {
		t.Fatal("expected sector label removed from quote right sidebar")
	}
	if strings.Contains(quoteView, "Capital Deck") || strings.Contains(quoteView, "Signal Engine") || strings.Contains(quoteView, "News Wire") {
		t.Fatal("expected removed dedicated tabs to stay absent")
	}
}

func TestNumberKeyFiveSelectsAITab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabMarkets

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	model = updated.(Model)

	if model.tabIdx != tabAI {
		t.Fatalf("expected key 5 to select AI tab, got %d", model.tabIdx)
	}
}

func TestOverviewDeskSparkUsesRightPanelWidth(t *testing.T) {
	series := []float64{190, 192, 194, 193, 197, 201, 205}
	spark := sparklineBlock(series, 24)
	if got := len([]rune(spark)); got != 24 {
		t.Fatalf("expected desk sparkline width 24, got %d", got)
	}
}
