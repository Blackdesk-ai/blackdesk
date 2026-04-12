package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func batchContainsMsg(cmd tea.Cmd, match func(tea.Msg) bool) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	if match(msg) {
		return true
	}
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return false
	}
	for _, subcmd := range batch {
		if batchContainsMsg(subcmd, match) {
			return true
		}
	}
	return false
}

func TestCommandPaletteFilingsOpensQuoteFilingsMode(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "filings", Title: "Filings"},
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterFilings {
		t.Fatalf("expected Quote filings mode, got tab=%d mode=%d", m.tabIdx, m.quoteCenterMode)
	}
	if cmd == nil {
		t.Fatal("expected filings load command")
	}
}

func TestQuoteFilingsViewRendersLoadedFilings(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.quote.Symbol = "AAPL"
	model.quote.Price = 210
	model.fundamentals.Symbol = "AAPL"
	model.fundamentals.FiftyTwoWeekHigh = 220
	model.fundamentals.FiftyTwoWeekLow = 160
	filings, err := filingsProvider{}.GetFilings(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("GetFilings error: %v", err)
	}
	model.filings = filings

	view := model.View()
	if !strings.Contains(view, "FILINGS") {
		t.Fatal("expected filings section")
	}
	if !strings.Contains(view, "Annual report") || !strings.Contains(view, "PREVIEW") {
		t.Fatal("expected filings list and preview")
	}
	if strings.Contains(view, "SYMBOLS") || strings.Contains(view, "\nNEWS\n") {
		t.Fatal("expected filings page to hide quote sidebars")
	}
	if !strings.Contains(view, "DATE") || !strings.Contains(view, "FORM") {
		t.Fatal("expected filings list to show date and form columns")
	}
}

func TestQuoteFilingsModeEnterUsesSelectedFilingURL(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.filings, _ = filingsProvider{}.GetFilings(context.Background(), "AAPL")
	calledURL := ""
	original := openURLFunc
	openURLFunc = func(raw string) error {
		calledURL = raw
		return nil
	}
	defer func() { openURLFunc = original }()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if calledURL == "" {
		t.Fatal("expected filing URL to open")
	}
	if !strings.Contains(m.status, "Opened SEC filing") {
		t.Fatalf("expected SEC filing status, got %q", m.status)
	}
}

func TestQuoteFilingsModeUsesArrowNavigationInsteadOfWatchlist(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL", "MSFT"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	filings, err := filingsProvider{}.GetFilings(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("GetFilings error: %v", err)
	}
	model.filings = filings

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(Model)
	if m.selectedIdx != 0 {
		t.Fatalf("expected watchlist selection unchanged, got %d", m.selectedIdx)
	}
	if m.filingsSel != 1 {
		t.Fatalf("expected filings selection to move, got %d", m.filingsSel)
	}
}

func TestQuoteFilingsModeIgnoresOpenNewsKey(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.news = []domain.NewsItem{{Title: "Story", URL: "https://example.com/story"}}
	model.newsSelected = 0

	calledURL := ""
	original := openURLFunc
	openURLFunc = func(raw string) error {
		calledURL = raw
		return nil
	}
	defer func() { openURLFunc = original }()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m := updated.(Model)
	if calledURL != "" {
		t.Fatal("expected open news key to be ignored in filings mode")
	}
	if strings.Contains(m.status, "Opened news item") {
		t.Fatalf("expected filings mode status to remain unchanged, got %q", m.status)
	}
}

func TestQuoteFilingsViewKeepsLongFormInsideFormColumn(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.width = 120
	model.height = 30
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.filings = domain.FilingsSnapshot{
		Symbol: "AAPL",
		Items: []domain.FilingItem{
			{Form: "SCHEDULE 13G/A", FilingDate: time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC)},
		},
	}

	view := model.View()
	if !strings.Contains(view, "SCHEDULE 13G/A") {
		t.Fatal("expected long SEC form code to stay visible in the form column")
	}
	if !strings.Contains(view, "SEC filing") {
		t.Fatal("expected meaning column to remain visible next to long SEC form code")
	}
}

func TestQuoteFilingsModeIStartsAIAnalysisFlow(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.filings, _ = filingsProvider{}.GetFilings(context.Background(), "AAPL")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m := updated.(Model)
	if m.tabIdx != tabAI {
		t.Fatalf("expected filings analysis key to switch to AI tab, got %d", m.tabIdx)
	}
	if !m.aiRunning {
		t.Fatal("expected filings analysis key to start AI thinking state immediately")
	}
	if len(m.aiMessages) == 0 || m.aiMessages[len(m.aiMessages)-1].Role != aiMessageUser {
		t.Fatal("expected filings analysis key to append an AI user prompt")
	}
	if cmd == nil {
		t.Fatal("expected filings analysis key to start preparation command")
	}

	nextMsg := cmd()
	prepared, ok := nextMsg.(aiFilingAnalysisPreparedMsg)
	if !ok {
		t.Fatalf("expected filing analysis prepared message, got %T", nextMsg)
	}
	if !strings.Contains(prepared.filing.Text, "Revenue grew 12%") {
		t.Fatalf("expected filing document text in prepared message, got %q", prepared.filing.Text)
	}
}

func TestQuoteFilingsFilterShowsOnlyPeriodicReports(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.filingsFilter = filingsFilterPeriodicReports
	model.filings = domain.FilingsSnapshot{
		Symbol: "AAPL",
		Items: []domain.FilingItem{
			{Form: "10-K", FilingDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
			{Form: "10-Q", FilingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
			{Form: "8-K", FilingDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
			{Form: "4", FilingDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	view := model.View()
	if !strings.Contains(view, "10-K") || !strings.Contains(view, "10-Q") {
		t.Fatal("expected periodic reports filter to keep 10-K and 10-Q visible")
	}
	if strings.Contains(view, "\n8-K") || strings.Contains(view, "\n4  ") {
		t.Fatal("expected periodic reports filter to hide non-periodic filings")
	}
	if !strings.Contains(view, "10-K/10-Q") {
		t.Fatal("expected current filings filter tab to be visible in the list header")
	}
}

func TestQuoteFilingsFilterTabsShowAllOptions(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings

	line := model.renderFilingsFilterTabs(lipgloss.NewStyle(), 46)
	for _, label := range []string{"ALL", "10-K", "10-Q", "10-K/10-Q"} {
		if !strings.Contains(line, label) {
			t.Fatalf("expected filings filter tabs to include %q, got %q", label, line)
		}
	}
}

func TestQuoteFilingsFilterCycleWithArrowKeys(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.filings = domain.FilingsSnapshot{
		Symbol: "AAPL",
		Items: []domain.FilingItem{
			{Form: "10-K", FilingDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
			{Form: "10-Q", FilingDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
			{Form: "8-K", FilingDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	m := updated.(Model)
	if m.filingsFilter != filingsFilterPeriodicReports {
		t.Fatalf("expected first filings filter step to select 10-K/10-Q, got %v", m.filingsFilter)
	}
	if item, ok := m.currentFiling(); !ok || (item.Form != "10-K" && item.Form != "10-Q") {
		t.Fatal("expected current filing selection to stay on a visible periodic report")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.filingsFilter != filingsFilterTenK {
		t.Fatalf("expected second filings filter step to select 10-K, got %v", m.filingsFilter)
	}
	if item, ok := m.currentFiling(); !ok || item.Form != "10-K" {
		t.Fatal("expected current filing selection to move to the visible 10-K item")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.filingsFilter != filingsFilterPeriodicReports {
		t.Fatalf("expected reverse filings filter step to return to 10-K/10-Q, got %v", m.filingsFilter)
	}
}

func TestSearchOpenSymbolInFilingsModeLoadsFilingsForNewSymbol(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		Registry:        providers.NewRegistry(testProvider{}),
		FilingsProvider: filingsProvider{},
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.searchMode = true
	model.searchInput.Focus()
	model.searchInput.SetValue("msft")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.config.ActiveSymbol != "MSFT" {
		t.Fatalf("expected active symbol MSFT, got %q", m.config.ActiveSymbol)
	}
	if cmd == nil {
		t.Fatal("expected search enter to trigger workspace reload command")
	}
	if !batchContainsMsg(cmd, func(msg tea.Msg) bool {
		_, ok := msg.(filingsLoadedMsg)
		return ok
	}) {
		t.Fatal("expected filings reload to be included when opening a new symbol from search in filings mode")
	}
}
