package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestCommandPaletteEarningsOpensQuoteEarningsMode(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "earnings", Title: "Earnings"},
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterEarnings {
		t.Fatalf("expected Quote earnings mode, got tab=%d mode=%d", m.tabIdx, m.quoteCenterMode)
	}
	if cmd == nil {
		t.Fatal("expected earnings load command")
	}
}

func TestQuoteEarningsViewRendersLoadedEarnings(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterEarnings
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.quote.Symbol = "AAPL"
	model.quote.Price = 210
	model.earnings = sampleEarningsSnapshot()

	view := model.View()
	if !strings.Contains(view, "EARNINGS") {
		t.Fatal("expected earnings section")
	}
	if !strings.Contains(view, "Next earnings") || !strings.Contains(view, "Estimate Trend") {
		t.Fatal("expected earnings list and preview")
	}
	if !strings.Contains(view, "Trend") || !strings.Contains(view, "EPS trend is rising") {
		t.Fatal("expected preview to show earnings trend summary")
	}
	if strings.Contains(view, "Next est") || strings.Contains(view, "Net +") || strings.Contains(view, "->") {
		t.Fatal("expected compact trend summary without path details")
	}
	if strings.Contains(view, "SYMBOLS") || strings.Contains(view, "\nNEWS\n") {
		t.Fatal("expected earnings page to hide quote sidebars")
	}
	if !strings.Contains(view, "DATE") || !strings.Contains(view, "TYPE") {
		t.Fatal("expected earnings list to show date and type columns")
	}
}

func TestQuoteEarningsModeUsesArrowNavigationInsteadOfWatchlist(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterEarnings
	model.config.Watchlist = []string{"AAPL", "MSFT"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.earnings = sampleEarningsSnapshot()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(Model)
	if m.selectedIdx != 0 {
		t.Fatalf("expected watchlist selection unchanged, got %d", m.selectedIdx)
	}
	if m.earningsSel != 1 {
		t.Fatalf("expected earnings selection to move, got %d", m.earningsSel)
	}
}

func TestEarningsTrendDirectionClassifiesRising(t *testing.T) {
	label, _ := earningsTrendDirection([]float64{0.42, 0.47, 0.53, 0.61}, lipgloss.NewStyle(), lipgloss.NewStyle())
	if label != "EPS trend is rising" {
		t.Fatalf("expected rising trend, got %q", label)
	}
}

func TestEarningsTrendDirectionClassifiesFalling(t *testing.T) {
	label, _ := earningsTrendDirection([]float64{0.61, 0.54, 0.41, 0.33}, lipgloss.NewStyle(), lipgloss.NewStyle())
	if label != "EPS trend is falling" {
		t.Fatalf("expected falling trend, got %q", label)
	}
}

func TestEarningsTrendDirectionClassifiesMixedWhenSeriesWhipsaws(t *testing.T) {
	label, _ := earningsTrendDirection([]float64{0.49, 0.53, 0.35, 0.12}, lipgloss.NewStyle(), lipgloss.NewStyle())
	if label != "EPS trend is mixed" {
		t.Fatalf("expected mixed trend, got %q", label)
	}
}

func TestEarningsTrendDirectionClassifiesFlatWhenUnchanged(t *testing.T) {
	label, _ := earningsTrendDirection([]float64{0.50, 0.50, 0.50}, lipgloss.NewStyle(), lipgloss.NewStyle())
	if label != "EPS trend is flat" {
		t.Fatalf("expected flat trend, got %q", label)
	}
}
