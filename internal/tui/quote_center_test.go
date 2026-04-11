package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestQuoteTabNavigationSwitchesCenterBetweenChartAndFundamentals(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.config.Watchlist = []string{"AAPL"}
	model.selectedIdx = 0
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
			{Time: time.Now().AddDate(0, 0, -3), Open: 195, High: 199, Low: 194, Close: 198},
			{Time: time.Now().AddDate(0, 0, -2), Open: 198, High: 204, Low: 197, Close: 202},
			{Time: time.Now().AddDate(0, 0, -1), Open: 202, High: 211, Low: 201, Close: 210},
		},
	}
	model.fundamentals = domain.FundamentalsSnapshot{
		Symbol:                  "AAPL",
		Sector:                  "Technology",
		Industry:                "Consumer Electronics",
		Description:             "Apple designs consumer devices and software.",
		MarketCap:               3_200_000_000_000,
		EnterpriseValue:         3_280_000_000_000,
		TrailingPE:              31.2,
		ForwardPE:               28.4,
		PEGRatio:                2.14,
		PriceToSales:            8.10,
		EnterpriseToRevenue:     8.30,
		EnterpriseToEBITDA:      24.5,
		BookValue:               4.25,
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
		DebtToEquity:            1.67,
		RecommendationMean:      1.9,
		RecommendationKey:       "buy",
		AnalystOpinions:         39,
		TargetLowPrice:          185,
		TargetMeanPrice:         225,
		TargetHighPrice:         250,
	}

	chartView := model.View()
	if !strings.Contains(chartView, "TIMEFRAMES") {
		t.Fatal("expected chart mode to render timeframe controls")
	}
	if !strings.Contains(ansi.Strip(chartView), "5Y ←/→") {
		t.Fatal("expected timeframe key hint after the last timeframe label")
	}
	if strings.Contains(chartView, "Fundamentals board") {
		t.Fatal("expected chart mode to hide fundamentals board hint")
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	model = updated.(Model)

	fundamentalsView := model.View()
	if strings.Contains(fundamentalsView, "TIMEFRAMES") {
		t.Fatal("expected fundamentals mode to replace chart section")
	}
	if strings.Contains(fundamentalsView, "Fundamentals board") {
		t.Fatal("expected fundamentals mode hint to be removed")
	}
	if !strings.Contains(fundamentalsView, "VALUATION") || !strings.Contains(fundamentalsView, "FINANCIALS") || !strings.Contains(fundamentalsView, "PROFITABILITY") {
		t.Fatal("expected fundamentals cards in quote center")
	}
	if !strings.Contains(fundamentalsView, "ROIC") {
		t.Fatal("expected ROIC to be shown in profitability card")
	}
	if !strings.Contains(fundamentalsView, "PROFITABILITY") {
		t.Fatal("expected implied eps growth band in profitability card")
	}
	if strings.Contains(fundamentalsView, "ANALYST VIEW") {
		t.Fatal("expected analyst view card to be removed from fundamentals board")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = updated.(Model)
	backToChart := model.View()
	if !strings.Contains(backToChart, "TIMEFRAMES") {
		t.Fatal("expected c to return to chart view")
	}
}

func TestQuoteTabNavigationShowsTechnicalsBoard(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
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
		Range:    "6mo",
		Interval: "1d",
	}
	for i := 90; i >= 0; i-- {
		price := 150.0 + float64(90-i)*0.8
		model.series.Candles = append(model.series.Candles, domain.Candle{
			Time:   time.Now().AddDate(0, 0, -i),
			Open:   price - 1.2,
			High:   price + 1.8,
			Low:    price - 2.1,
			Close:  price,
			Volume: 60_000_000 + int64(i*120_000),
		})
	}
	model.technicalCache["AAPL"] = model.series

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updated.(Model)

	technicalsView := model.View()
	if strings.Contains(technicalsView, "TIMEFRAMES") {
		t.Fatal("expected technicals mode to replace chart section")
	}
	if strings.Contains(technicalsView, "Technical board") {
		t.Fatal("expected technical board hint to be removed")
	}
	if !strings.Contains(technicalsView, "MOMENTUM") || !strings.Contains(technicalsView, "TREND") {
		t.Fatal("expected momentum and trend cards in quote center")
	}
	if !strings.Contains(technicalsView, "VOLATILITY") || !strings.Contains(technicalsView, "STAT EDGE") {
		t.Fatal("expected volatility and stats cards in quote center")
	}
	if !strings.Contains(technicalsView, "PriceZ 21") {
		t.Fatal("expected PriceZ metric in technical board")
	}
	if !strings.Contains(technicalsView, "ADX 14") {
		t.Fatal("expected ADX metric in technical board")
	}
	if !strings.Contains(technicalsView, "ATR 14") {
		t.Fatal("expected ATR metric in technical board")
	}
	if !strings.Contains(technicalsView, "HV Rank") {
		t.Fatal("expected HV Rank metric in technical board")
	}
	if !strings.Contains(technicalsView, "HV Pctl") {
		t.Fatal("expected HV Percentile metric in technical board")
	}
	if !strings.Contains(technicalsView, "Signal") {
		t.Fatal("expected technical board to keep signal column")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = updated.(Model)
	backToChart := model.View()
	if !strings.Contains(backToChart, "TIMEFRAMES") {
		t.Fatal("expected c to return to chart view")
	}
}

func TestQuoteCenterMetricLabelsUseAnalystsLabelStyle(t *testing.T) {
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8C9B8"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F3EBDD"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#9F907E"))

	fundamentalsCard := renderQuoteFundamentalsCard(sectionStyle, labelStyle, muted, 28, []marketTableRow{
		{name: "Market cap", price: "3.2T"},
	}, "VALUATION")
	if !strings.Contains(fundamentalsCard, labelStyle.Render("Market cap")) {
		t.Fatal("expected fundamentals metric labels to use analysts label style")
	}

	splitCard := renderQuoteFundamentalsSplitCard(sectionStyle, labelStyle, muted, 42, []marketTableRow{
		{name: "Gross margin", price: "46.20%"},
		{name: "Operating", price: "31.10%"},
	}, []marketTableRow{
		{name: "ROIC", price: "42.10%"},
		{name: "Growth Est.", price: "13%-15%"},
	}, "PROFITABILITY")
	if strings.Count(splitCard, "Name") != 2 || strings.Count(splitCard, "Value") != 2 {
		t.Fatal("expected split fundamentals card to render two table headers")
	}
	if !strings.Contains(splitCard, "PROFITABILITY") {
		t.Fatal("expected split fundamentals card to include implied eps row")
	}

	technicalsCard := renderQuoteTechnicalsCard(sectionStyle, labelStyle, muted, 28, []marketTableRow{
		{name: "RSI 14", price: "62.1", chg: "bull", move: 1, styled: true},
	}, "MOMENTUM")
	if !strings.Contains(technicalsCard, labelStyle.Render("RSI 14")) {
		t.Fatal("expected technical metric labels to use analysts label style")
	}

	statementRow := renderStatementRow(domain.FinancialStatement{
		Frequency: domain.StatementFrequencyAnnual,
		Periods: []domain.StatementPeriod{
			{Label: "FY 2024"},
		},
	}, domain.StatementRow{
		Label: "Operating Cash Flow",
		Values: []domain.StatementValue{
			{Value: 115_800_000_000, Present: true},
		},
	}, 1, len("Operating Cash Flow"), len("115.80B"), labelStyle)
	if !strings.Contains(statementRow, labelStyle.Render("Operating Cash Flow")) {
		t.Fatal("expected statement row labels to use analysts label style")
	}
}

func TestRenderQuoteFundamentalsGridStacksCardsWhenWidthIsNarrow(t *testing.T) {
	quote := domain.QuoteSnapshot{Symbol: "AAPL"}
	fundamentals := domain.FundamentalsSnapshot{
		TrailingPE:              31.2,
		ForwardPE:               28.4,
		PEGRatio:                2.14,
		PriceToBook:             45.1,
		PriceToSales:            8.10,
		EnterpriseToEBITDA:      24.5,
		GrossMargins:            0.462,
		ProfitMargins:           0.262,
		OperatingMargins:        0.311,
		ReturnOnAssets:          0.284,
		ReturnOnEquity:          1.55,
		ReturnOnInvestedCapital: 0.421,
		RevenueGrowth:           0.061,
		EarningsGrowth:          0.104,
		Revenue:                 391_000_000_000,
		EBITDA:                  134_000_000_000,
		OperatingCashflow:       122_000_000_000,
		FreeCashflow:            99_000_000_000,
		TotalCash:               67_000_000_000,
		TotalDebt:               110_000_000_000,
		CurrentRatio:            0.99,
		RevenuePerShare:         25.1,
	}

	view := renderQuoteFundamentalsGrid(lipgloss.NewStyle().Bold(true), lipgloss.NewStyle().Bold(true), lipgloss.NewStyle(), quote, fundamentals, 56, 40)
	if strings.Count(view, "VALUATION") != 1 || strings.Count(view, "PROFITABILITY") != 1 || strings.Count(view, "FINANCIALS") == 0 {
		t.Fatal("expected stacked fundamentals cards to keep all sections visible")
	}
	if !strings.Contains(view, "Op cash flow") || !strings.Contains(view, "Free cash flow") {
		t.Fatal("expected narrow stacked layout to keep lower financial rows visible")
	}
}

func TestRenderQuoteFundamentalsGridUsesSingleSplitFinancialsCardOnWideLayout(t *testing.T) {
	quote := domain.QuoteSnapshot{Symbol: "AAPL", Price: 210}
	fundamentals := domain.FundamentalsSnapshot{
		TrailingPE:              31.2,
		ForwardPE:               28.4,
		PEGRatio:                2.14,
		PriceToBook:             45.1,
		PriceToSales:            8.10,
		EnterpriseToEBITDA:      24.5,
		DividendYield:           0.0048,
		GrossMargins:            0.462,
		ProfitMargins:           0.262,
		OperatingMargins:        0.311,
		ReturnOnAssets:          0.284,
		ReturnOnEquity:          1.55,
		ReturnOnInvestedCapital: 0.421,
		RevenueGrowth:           0.061,
		EarningsGrowth:          0.104,
		Revenue:                 391_000_000_000,
		EBITDA:                  134_000_000_000,
		OperatingCashflow:       122_000_000_000,
		FreeCashflow:            99_000_000_000,
		TotalCash:               67_000_000_000,
		TotalDebt:               110_000_000_000,
		CurrentRatio:            0.99,
		RevenuePerShare:         25.1,
	}

	view := renderQuoteFundamentalsGrid(lipgloss.NewStyle().Bold(true), lipgloss.NewStyle().Bold(true), lipgloss.NewStyle(), quote, fundamentals, 100, 60)
	if strings.Count(view, "FINANCIALS") != 1 {
		t.Fatal("expected wide fundamentals layout to render a single financials card")
	}
	if strings.Count(view, "Name") != 5 || strings.Count(view, "Value") != 5 {
		t.Fatal("expected wide fundamentals layout to render valuation plus two split cards")
	}
	if !strings.Contains(view, "Earnings Yield") {
		t.Fatal("expected valuation card to include earnings yield")
	}
}

func TestQuoteTabNavigationShowsStatementsBoardWhenSupported(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(statementsProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quote = domain.QuoteSnapshot{
		Symbol:            "AAPL",
		ShortName:         "Apple",
		Currency:          "USD",
		Price:             210.12,
		Change:            3.15,
		ChangePercent:     1.52,
		MarketState:       domain.MarketStateRegular,
		RegularMarketTime: time.Now(),
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected statements load command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "Income") || !strings.Contains(view, "Balance Sheet") || !strings.Contains(view, "Cash Flow") {
		t.Fatal("expected internal statement kind tabs")
	}
	if !strings.Contains(view, "←/→") {
		t.Fatal("expected statement kind key hint in header")
	}
	if !strings.Contains(view, "Annual") || !strings.Contains(view, "Quarterly") {
		t.Fatal("expected internal statement frequency tabs")
	}
	if !strings.Contains(view, "[/]") {
		t.Fatal("expected statement frequency key hint in header")
	}
	if !strings.Contains(view, "Total Revenue") || !strings.Contains(view, "Net Income") {
		t.Fatal("expected statement rows")
	}
	if !strings.Contains(view, "(+2.0%)") || !strings.Contains(view, "(+4.0%)") {
		t.Fatal("expected latest statement column to include growth versus prior period")
	}
	if !strings.Contains(view, "FY 2024") || !strings.Contains(view, "FY 2023") {
		t.Fatal("expected statement periods")
	}
	if strings.Contains(view, "TIMEFRAMES") {
		t.Fatal("expected statements mode to replace chart section")
	}
	if strings.Contains(view, "NEWS") || strings.Contains(view, "PROFILE") {
		t.Fatal("expected statements mode to hide bottom panels")
	}
	if !strings.Contains(view, "SYMBOLS") || !strings.Contains(view, "MARKET HEAT") {
		t.Fatal("expected left and right sidebars to remain visible")
	}
}

func TestQuoteTabNavigationShowsInsidersBoardWhenSupported(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(insidersProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quote = domain.QuoteSnapshot{
		Symbol:            "AAPL",
		ShortName:         "Apple",
		Currency:          "USD",
		Price:             210.12,
		Change:            3.15,
		ChangePercent:     1.52,
		MarketState:       domain.MarketStateRegular,
		RegularMarketTime: time.Now(),
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected insiders load command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "INSIDER FLOW") || !strings.Contains(view, "RECENT TRANSACTIONS") {
		t.Fatal("expected insider summary and transaction cards")
	}
	if strings.Contains(view, "INSIDER ROSTER") {
		t.Fatal("expected insider roster to be removed from the default view")
	}
	if !strings.Contains(view, "Buy shares") || !strings.Contains(view, "Net shares") {
		t.Fatal("expected insider activity summary rows")
	}
	if !strings.Contains(view, "SIEFFERT") || !strings.Contains(view, "THORNOCK") {
		t.Fatal("expected insider names in board")
	}
	if strings.Contains(view, "Type") || strings.Contains(view, "Details") {
		t.Fatal("expected insiders board to keep the compact transaction layout")
	}
	if strings.Contains(view, "TIMEFRAMES") {
		t.Fatal("expected insiders mode to replace chart section")
	}
	if strings.Contains(view, "NEWS") || strings.Contains(view, "PROFILE") {
		t.Fatal("expected insiders mode to hide bottom panels")
	}
}
