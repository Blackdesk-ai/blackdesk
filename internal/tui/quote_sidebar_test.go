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

func TestViewShowsWatchlistQuotesInSidebar(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.width = 140
	model.height = 36
	model.watchQuotes["AAPL"] = domain.QuoteSnapshot{Symbol: "AAPL", Price: 210.12, ChangePercent: 1.52}
	model.watchQuotes["MSFT"] = domain.QuoteSnapshot{Symbol: "MSFT", Price: 498.44, ChangePercent: -0.81}

	view := model.View()
	if !strings.Contains(view, "SYMBOLS") {
		t.Fatal("expected symbols sidebar")
	}
	if !strings.Contains(view, "↑/↓") {
		t.Fatal("expected watchlist navigation hint in symbols header")
	}
	if !strings.Contains(view, "210.12") {
		t.Fatal("expected active symbol price in sidebar")
	}
	if !strings.Contains(view, "+1.52%") {
		t.Fatal("expected active symbol move in sidebar")
	}
	if !strings.Contains(view, "▲ +1.52%") {
		t.Fatal("expected positive arrow in sidebar")
	}
	if !strings.Contains(view, "498.44") {
		t.Fatal("expected secondary symbol price in sidebar")
	}
	if !strings.Contains(view, "-0.81%") {
		t.Fatal("expected secondary symbol move in sidebar")
	}
	if !strings.Contains(view, "▼ -0.81%") {
		t.Fatal("expected negative arrow in sidebar")
	}
}

func TestViewShowsPremarketWatchlistQuotesInSidebar(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.width = 140
	model.height = 36
	model.watchQuotes["AAPL"] = domain.QuoteSnapshot{
		Symbol:              "AAPL",
		Price:               210.12,
		ChangePercent:       1.52,
		MarketState:         domain.MarketStatePre,
		PreMarketPrice:      214.55,
		PreMarketChangePerc: 3.41,
	}
	model.watchQuotes["MSFT"] = domain.QuoteSnapshot{
		Symbol:               "MSFT",
		Price:                498.44,
		ChangePercent:        -0.81,
		MarketState:          domain.MarketStatePost,
		PostMarketPrice:      496.10,
		PostMarketChangePerc: -1.28,
	}

	view := model.View()
	if !strings.Contains(view, "214.55") {
		t.Fatal("expected premarket price in sidebar")
	}
	if !strings.Contains(view, "▲ +3.41%") {
		t.Fatal("expected premarket move in sidebar")
	}
	if !strings.Contains(view, "496.10") {
		t.Fatal("expected postmarket price in sidebar")
	}
	if !strings.Contains(view, "▼ -1.28%") {
		t.Fatal("expected postmarket move in sidebar")
	}
}

func TestQuoteInsightBlockShowsThinkingAndCachedText(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.config.Watchlist = []string{"AAPL"}
	model.selectedIdx = 0
	model.config.ActiveSymbol = "AAPL"
	model.aiQuoteInsightSymbol = "AAPL"
	model.aiQuoteInsightRunning = true
	model.clock = time.Date(2026, 4, 3, 12, 0, 2, 0, time.UTC)

	block := ansi.Strip(model.renderQuoteInsightBlock(lipgloss.NewStyle(), 36))
	if !strings.Contains(block, "thinking") {
		t.Fatal("expected thinking indicator while quote insight runs")
	}

	model.aiQuoteInsightRunning = false
	model.aiQuoteInsight = ""
	model.aiQuoteInsightUpdated = time.Time{}
	block = ansi.Strip(model.renderQuoteInsightBlock(lipgloss.NewStyle(), 36))
	if !strings.Contains(block, "Press i") || !strings.Contains(block, "generate AI insight") {
		t.Fatal("expected quote insight empty-state hint")
	}

	model.aiQuoteInsight = "Buy: trend and target support remain constructive, though valuation is full."
	model.aiQuoteInsightUpdated = time.Date(2026, 4, 3, 12, 4, 0, 0, time.UTC)

	block = ansi.Strip(model.renderQuoteInsightBlock(lipgloss.NewStyle(), 36))
	if !strings.Contains(block, "Buy: trend and target support") {
		t.Fatal("expected cached quote insight text")
	}
	if !strings.Contains(block, "Updated ") {
		t.Fatal("expected quote insight updated timestamp")
	}
}

func TestQuoteViewShowsAIInsightBlock(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 42
	model.tabIdx = tabQuote
	model.config.Watchlist = []string{"AAPL"}
	model.selectedIdx = 0
	model.quote = domain.QuoteSnapshot{Symbol: "AAPL", Price: 210}
	model.fundamentals = domain.FundamentalsSnapshot{Symbol: "AAPL", Sector: "Technology", ForwardPE: 28.4, EPS: 4.9, Beta: 1.18}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "AI INSIGHT") {
		t.Fatal("expected quote AI insight block in sidebar")
	}
	if !strings.Contains(view, "AI INSIGHT (i)") {
		t.Fatal("expected quote AI insight header key hint in sidebar")
	}
}

func TestProfilePanelShowsSectorInHeader(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.fundamentals = domain.FundamentalsSnapshot{
		Sector:      "Technology",
		Description: "Apple designs consumer devices and software.",
	}

	panel := ansi.Strip(model.renderProfilePanel(lipgloss.NewStyle().Bold(true), lipgloss.NewStyle(), 36, 8))
	if !strings.Contains(panel, "PROFILE (p)") {
		t.Fatal("expected profile header key hint")
	}
	if !strings.Contains(panel, "Technology") {
		t.Fatal("expected sector value in profile header")
	}
}

func TestQuoteRefreshAndInsightKeysAreSeparated(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(Model)
	if model.aiQuoteInsightRunning {
		t.Fatal("expected refresh key to leave quote AI insight idle")
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	if !model.aiQuoteInsightRunning {
		t.Fatal("expected insight key to start quote AI insight refresh")
	}
	if cmd == nil {
		t.Fatal("expected insight key to prepare quote insight context")
	}
}

func TestRenderLeftPanelUsesFullHeightAndShowsOverflowIndicators(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.selectedIdx = 5
	model.watchlistScroll = 2
	height := 10

	panel := model.renderLeftPanel(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		30,
		height,
	)

	lines := splitLines(panel)
	if len(lines) != height {
		t.Fatalf("expected panel height %d, got %d", height, len(lines))
	}
	if !strings.Contains(panel, "↑ more") {
		t.Fatal("expected top overflow indicator")
	}
	if !strings.Contains(panel, "↓ more") {
		t.Fatal("expected bottom overflow indicator")
	}
	if !strings.Contains(panel, "AAPL") {
		t.Fatal("expected list body to use available height for symbols")
	}
}

func TestWatchlistNavigationUpdatesVisualScroll(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.width = 140
	model.height = 18

	for range 10 {
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updated.(Model)
	}

	if model.watchlistScroll == 0 {
		t.Fatal("expected watchlist scroll to move with selection")
	}

	panel := model.renderLeftPanel(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		30,
		model.leftPanelContentHeight(),
	)
	if !strings.Contains(panel, "▶ TSLA") {
		t.Fatal("expected selected symbol to remain visible in rendered panel")
	}
}

func TestAddToWatchlistEvictsOldestWhenLimitExceeded(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.config.Watchlist = nil
	for i := 0; i < storage.MaxWatchlistItems; i++ {
		model.config.Watchlist = append(model.config.Watchlist, strings.ToUpper(string(rune('A'+(i%26))))+strings.Repeat("X", 3))
	}
	oldest := model.config.Watchlist[len(model.config.Watchlist)-1]

	model.addToWatchlist("TSLA")

	if got := len(model.config.Watchlist); got != storage.MaxWatchlistItems {
		t.Fatalf("expected watchlist capped at %d, got %d", storage.MaxWatchlistItems, got)
	}
	if model.config.Watchlist[0] != "TSLA" {
		t.Fatalf("expected newest symbol at front, got %s", model.config.Watchlist[0])
	}
	for _, item := range model.config.Watchlist {
		if item == oldest {
			t.Fatalf("expected oldest symbol %s to be evicted", oldest)
		}
	}
}

func TestAddToWatchlistSelectsExistingSymbol(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 18
	model.selectedIdx = 0
	model.watchlistScroll = 0

	model.addToWatchlist("TSLA")

	if model.selectedIdx != 10 {
		t.Fatalf("expected existing TSLA index 10, got %d", model.selectedIdx)
	}
	if model.config.ActiveSymbol != "TSLA" {
		t.Fatalf("expected active symbol TSLA, got %s", model.config.ActiveSymbol)
	}
	if model.watchlistScroll == 0 {
		t.Fatal("expected scroll to move to existing selected symbol")
	}
}

func TestDeleteWatchlistKeyOnlyWorksOnQuoteTab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabMarkets
	model.selectedIdx = 1
	before := append([]string(nil), model.config.Watchlist...)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(Model)

	if len(model.config.Watchlist) != len(before) {
		t.Fatalf("expected watchlist length unchanged off quote tab, got %d want %d", len(model.config.Watchlist), len(before))
	}
	for i := range before {
		if model.config.Watchlist[i] != before[i] {
			t.Fatalf("expected watchlist item %d to stay %q, got %q", i, before[i], model.config.Watchlist[i])
		}
	}
}

func TestQuoteOnlyKeysAreIgnoredOutsideQuoteTab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabMarkets
	model.news = []domain.NewsItem{{Title: "News", URL: "https://example.com/story"}}
	model.newsSelected = 0
	model.fundamentals.Description = "Long company profile text."
	model.selectedIdx = 1
	beforeWatchlist := append([]string(nil), model.config.Watchlist...)
	beforeStatus := model.status

	called := false
	prevOpenURL := openURLFunc
	openURLFunc = func(raw string) error {
		called = true
		return nil
	}
	defer func() {
		openURLFunc = prevOpenURL
	}()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updated.(Model)
	if model.newsSelected != 0 {
		t.Fatalf("expected news selection unchanged off quote tab, got %d", model.newsSelected)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updated.(Model)
	if model.profileScroll != 0 {
		t.Fatalf("expected profile scroll unchanged off quote tab, got %d", model.profileScroll)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	model = updated.(Model)
	if called {
		t.Fatal("expected open news key to be ignored off quote tab")
	}
	if model.status != beforeStatus {
		t.Fatalf("expected status unchanged off quote tab, got %q want %q", model.status, beforeStatus)
	}

	if len(model.config.Watchlist) != len(beforeWatchlist) {
		t.Fatalf("expected watchlist length unchanged off quote tab, got %d want %d", len(model.config.Watchlist), len(beforeWatchlist))
	}
	for i := range beforeWatchlist {
		if model.config.Watchlist[i] != beforeWatchlist[i] {
			t.Fatalf("expected watchlist item %d to stay %q, got %q", i, beforeWatchlist[i], model.config.Watchlist[i])
		}
	}
}

func TestOpenNewsKeyWorksOnlyWhenQuoteNewsPanelIsVisible(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterChart
	model.news = []domain.NewsItem{{Title: "News", URL: "https://example.com/story"}}
	model.newsSelected = 0

	calledURL := ""
	prevOpenURL := openURLFunc
	openURLFunc = func(raw string) error {
		calledURL = raw
		return nil
	}
	defer func() {
		openURLFunc = prevOpenURL
	}()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	model = updated.(Model)
	if calledURL != "https://example.com/story" {
		t.Fatalf("expected open news key to open visible quote news story, got %q", calledURL)
	}

	model.quoteCenterMode = quoteCenterStatements
	model.status = ""
	calledURL = ""

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	model = updated.(Model)
	if calledURL != "" {
		t.Fatal("expected open news key to be ignored when quote news panel is hidden")
	}
	if strings.Contains(model.status, "Opened news item") {
		t.Fatalf("expected no news-open status when quote news panel is hidden, got %q", model.status)
	}
}

func TestQuoteLocalPanelKeysAreIgnoredWhenPanelsAreHidden(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterFilings
	model.config.Watchlist = []string{"AAPL", "MSFT"}
	model.selectedIdx = 0
	model.config.ActiveSymbol = "AAPL"
	model.news = []domain.NewsItem{{Title: "Story", URL: "https://example.com/story"}, {Title: "Story 2", URL: "https://example.com/story-2"}}
	model.newsSelected = 0
	model.fundamentals.Description = strings.Repeat("profile text ", 50)
	model.profileScroll = 0

	calledURL := ""
	prevOpenURL := openURLFunc
	openURLFunc = func(raw string) error {
		calledURL = raw
		return nil
	}
	defer func() {
		openURLFunc = prevOpenURL
	}()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	if model.aiQuoteInsightRunning {
		t.Fatal("expected insight key to be ignored when quote panels are hidden")
	}

	beforeWatchlist := append([]string(nil), model.config.Watchlist...)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(Model)
	if len(model.config.Watchlist) != len(beforeWatchlist) {
		t.Fatal("expected delete key to be ignored when quote panels are hidden")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updated.(Model)
	if model.newsSelected != 0 {
		t.Fatal("expected news navigation key to be ignored when quote panels are hidden")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updated.(Model)
	if model.profileScroll != 0 {
		t.Fatal("expected profile key to be ignored when quote panels are hidden")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	model = updated.(Model)
	if calledURL != "" {
		t.Fatal("expected open news key to be ignored when quote panels are hidden")
	}
}
