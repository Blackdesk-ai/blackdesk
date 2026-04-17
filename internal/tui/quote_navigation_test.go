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

func TestClosedQuoteChartDoesNotRepeatClosedStateBadges(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quote = domain.QuoteSnapshot{
		Symbol:        "MSFT",
		ShortName:     "Microsoft",
		Currency:      "USD",
		Price:         366.00,
		Change:        -1.25,
		ChangePercent: -0.34,
		Volume:        94_320,
		MarketState:   domain.MarketStateClosed,
	}
	model.series = domain.PriceSeries{
		Symbol:   "MSFT",
		Range:    "1mo",
		Interval: "1d",
		Candles: []domain.Candle{
			{Time: time.Now().AddDate(0, 0, -2), Close: 370},
			{Time: time.Now().AddDate(0, 0, -1), Close: 368},
			{Time: time.Now(), Close: 366},
		},
	}
	model.fundamentals = domain.FundamentalsSnapshot{
		Symbol:           "MSFT",
		FiftyTwoWeekLow:  300,
		FiftyTwoWeekHigh: 420,
	}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "CLOSED") {
		t.Fatal("expected closed market badge")
	}
	if strings.Contains(view, "Last close") {
		t.Fatal("expected redundant last close label to be omitted when market is closed")
	}
}

func TestQuoteStatusShowsTimeframeKeysOnlyInChartMode(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote

	chartStatus := ansi.Strip(model.statusText())
	if strings.Contains(chartStatus, "←/→ timeframe") {
		t.Fatal("expected chart mode status to hide timeframe navigation after moving it into the panel header")
	}
	if strings.Contains(chartStatus, "↑/↓ symbols") {
		t.Fatal("expected chart mode status to hide watchlist navigation after moving it into the panel header")
	}
	if strings.Contains(chartStatus, "insight") {
		t.Fatal("expected chart mode status to hide quote insight shortcut after moving it into the sidebar header")
	}
	if strings.Contains(chartStatus, "news") || strings.Contains(chartStatus, "profile") || strings.Contains(chartStatus, "open news") {
		t.Fatal("expected chart mode status to hide bottom panel shortcuts after moving them into panel headers")
	}
	if strings.Contains(chartStatus, "chart") {
		t.Fatal("expected chart mode status to hide active chart shortcut")
	}
	if !strings.Contains(chartStatus, "fundamentals") || !strings.Contains(chartStatus, "technicals") {
		t.Fatal("expected chart mode status to show other quote shortcuts")
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	model = updated.(Model)

	fundamentalsStatus := ansi.Strip(model.statusText())
	if strings.Contains(fundamentalsStatus, "←/→ timeframe") {
		t.Fatal("expected fundamentals mode status to hide timeframe navigation")
	}
	if strings.Contains(fundamentalsStatus, "↑/↓ symbols") {
		t.Fatal("expected fundamentals mode status to hide watchlist navigation")
	}
	if strings.Contains(fundamentalsStatus, "insight") {
		t.Fatal("expected fundamentals mode status to hide quote insight shortcut")
	}
	if strings.Contains(fundamentalsStatus, "fundamentals") {
		t.Fatal("expected fundamentals mode status to hide active fundamentals shortcut")
	}
	if !strings.Contains(fundamentalsStatus, "chart") || !strings.Contains(fundamentalsStatus, "technicals") {
		t.Fatal("expected fundamentals mode status to show other quote shortcuts")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updated.(Model)

	technicalsStatus := ansi.Strip(model.statusText())
	if strings.Contains(technicalsStatus, "↑/↓ symbols") {
		t.Fatal("expected technicals mode status to hide watchlist navigation")
	}
	if strings.Contains(technicalsStatus, "insight") {
		t.Fatal("expected technicals mode status to hide quote insight shortcut")
	}
	if strings.Contains(technicalsStatus, "technicals") {
		t.Fatal("expected technicals mode status to hide active technicals shortcut")
	}
	if !strings.Contains(technicalsStatus, "chart") || !strings.Contains(technicalsStatus, "fundamentals") {
		t.Fatal("expected technicals mode status to show other quote shortcuts")
	}
}

func TestStatementsStatusHidesBottomPanelKeys(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(statementsProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterStatements

	status := ansi.Strip(model.statusText())
	if strings.Contains(status, "news") || strings.Contains(status, "profile") || strings.Contains(status, "open news") {
		t.Fatal("expected statements mode to hide bottom-panel quote shortcuts")
	}
	if !strings.Contains(status, "chart") || !strings.Contains(status, "fundamentals") || !strings.Contains(status, "technicals") {
		t.Fatal("expected statements mode to keep center navigation shortcuts")
	}
	if strings.Contains(status, "←/→ statement") || strings.Contains(status, "[/] frequency") {
		t.Fatal("expected statements mode to hide statement navigation shortcuts after moving them into the header")
	}
}

func TestInsidersStatusHidesBottomPanelKeysAndActiveShortcut(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(researchProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterInsiders

	status := ansi.Strip(model.statusText())
	if strings.Contains(status, "news") || strings.Contains(status, "profile") || strings.Contains(status, "open news") {
		t.Fatal("expected insiders mode to hide bottom-panel quote shortcuts")
	}
	if strings.Contains(status, "insiders") {
		t.Fatal("expected insiders mode status to hide active insiders shortcut")
	}
	if !strings.Contains(status, "chart") || !strings.Contains(status, "fundamentals") || !strings.Contains(status, "technicals") || !strings.Contains(status, "statements") {
		t.Fatal("expected insiders mode to keep other quote research shortcuts")
	}
}

func TestArrowKeysChangeTimeframeInQuoteChartAndSharpe(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	startRange := model.rangeIdx

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	if model.rangeIdx != (startRange+1)%len(ranges) {
		t.Fatalf("expected right arrow to advance timeframe from %d to %d, got %d", startRange, (startRange+1)%len(ranges), model.rangeIdx)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	model = updated.(Model)
	chartlessRange := model.rangeIdx

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)
	if model.rangeIdx != chartlessRange {
		t.Fatalf("expected left arrow to leave timeframe unchanged outside chart mode, got %d from %d", model.rangeIdx, chartlessRange)
	}

	model.quoteCenterMode = quoteCenterSharpe
	chartRangeBeforeSharpe := model.rangeIdx
	sharpeRange := model.sharpeRangeIdx
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)
	if model.sharpeRangeIdx == sharpeRange {
		t.Fatalf("expected left arrow to change sharpe timeframe, stayed at %d", model.sharpeRangeIdx)
	}
	if model.rangeIdx != chartRangeBeforeSharpe {
		t.Fatalf("expected chart timeframe unchanged while sharpe is visible, got %d from %d", model.rangeIdx, chartRangeBeforeSharpe)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	model = updated.(Model)
	quoteRange := model.rangeIdx

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)
	if model.rangeIdx != quoteRange {
		t.Fatalf("expected left arrow to leave timeframe unchanged on non-quote tab, got %d from %d", model.rangeIdx, quoteRange)
	}
}

func TestStatementsModeArrowKeysCycleStatementKinds(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(statementsProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterStatements
	model.statementKind = domain.StatementKindIncome

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model = updated.(Model)
	if model.statementKind != domain.StatementKindBalanceSheet {
		t.Fatalf("expected right arrow to move to balance sheet, got %q", model.statementKind)
	}
	if cmd == nil {
		t.Fatal("expected statement reload command on right arrow")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model = updated.(Model)
	if model.statementKind != domain.StatementKindIncome {
		t.Fatalf("expected left arrow to move back to income, got %q", model.statementKind)
	}
	if cmd == nil {
		t.Fatal("expected statement reload command on left arrow")
	}
}

func TestStatementsModeBracketKeysCycleFrequency(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(statementsProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterStatements
	model.statementFreq = domain.StatementFrequencyAnnual

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(Model)
	if model.statementFreq != domain.StatementFrequencyQuarterly {
		t.Fatalf("expected ] to switch to quarterly, got %q", model.statementFreq)
	}
	if cmd == nil {
		t.Fatal("expected statement reload command on ]")
	}

	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = updated.(Model)
	if model.statementFreq != domain.StatementFrequencyAnnual {
		t.Fatalf("expected [ to switch back to annual, got %q", model.statementFreq)
	}
	if cmd == nil {
		t.Fatal("expected statement reload command on [")
	}
}

func TestNewsKeyCyclesSelectionFromTop(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.news = []domain.NewsItem{{Title: "one"}, {Title: "two"}, {Title: "three"}}
	model.newsSelected = 2

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updated.(Model)

	if model.newsSelected != 0 {
		t.Fatalf("expected news selection to wrap to 0, got %d", model.newsSelected)
	}
}

func TestProfileKeyCyclesScrollBackToTop(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.width = 140
	model.height = 20
	model.fundamentals.Description = strings.Repeat("profile text ", 200)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updated.(Model)
	if model.profileScroll == 0 {
		t.Fatal("expected first p to advance profile scroll")
	}

	model.profileScroll = 1 << 20
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updated.(Model)
	if model.profileScroll != 0 {
		t.Fatalf("expected profile scroll to wrap to top, got %d", model.profileScroll)
	}
}

func TestTechnicalHistoryLoadsTwoYearsOncePerSession(t *testing.T) {
	provider := &countingHistoryProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.tabIdx = tabQuote
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 4

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected first technical navigation to request 2y history")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if len(provider.historyCalls) != 1 {
		t.Fatalf("expected one 2y history request, got %d", len(provider.historyCalls))
	}
	if provider.historyCalls[0] != "AAPL|2y|1d" {
		t.Fatalf("unexpected history request %q", provider.historyCalls[0])
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	model = updated.(Model)
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	model = updated.(Model)
	if cmd != nil {
		t.Fatal("expected cached technical history to avoid a second 2y request")
	}
	if len(provider.historyCalls) != 1 {
		t.Fatalf("expected cached history to be reused, got %d calls", len(provider.historyCalls))
	}
}
