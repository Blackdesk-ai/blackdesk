package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestCommandPaletteAnalystOpensQuoteAnalystMode(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "analyst", Title: "Analyst Recommendations"},
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterAnalyst {
		t.Fatalf("expected Quote analyst mode, got tab=%d mode=%d", m.tabIdx, m.quoteCenterMode)
	}
	if cmd == nil {
		t.Fatal("expected analyst load command")
	}
}

func TestQuoteAnalystViewRendersLatestRecommendationsByDate(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterAnalyst
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.quote.Symbol = "AAPL"
	model.quote.Price = 210
	model.analyst = sampleAnalystRecommendationsSnapshot()

	view := model.View()
	if !strings.Contains(view, "ANALYST RECOMMENDATIONS") {
		t.Fatal("expected analyst recommendations section")
	}
	if !strings.Contains(view, "Morgan Stanley") || !strings.Contains(view, "Consensus") {
		t.Fatal("expected analyst list and preview")
	}
	if !strings.Contains(strings.ToLower(view), "distribution looks") {
		t.Fatal("expected trend interpretation to explain the monthly distribution")
	}
	if strings.Contains(view, "SYMBOLS") || strings.Contains(view, "\nNEWS\n") {
		t.Fatal("expected analyst page to hide quote sidebars")
	}
	first := strings.Index(view, "Morgan Stanley")
	second := strings.Index(view, "Goldman Sachs")
	if first == -1 || second == -1 || first > second {
		t.Fatal("expected latest analyst updates to be ordered by descending date")
	}
}

func TestQuoteAnalystModeUsesArrowNavigationInsteadOfWatchlist(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterAnalyst
	model.config.Watchlist = []string{"AAPL", "MSFT"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.analyst = sampleAnalystRecommendationsSnapshot()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(Model)
	if m.selectedIdx != 0 {
		t.Fatalf("expected watchlist selection unchanged, got %d", m.selectedIdx)
	}
	if m.analystSel != 1 {
		t.Fatalf("expected analyst selection to move, got %d", m.analystSel)
	}
}
