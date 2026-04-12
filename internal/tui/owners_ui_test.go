package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestCommandPaletteOwnersOpensQuoteOwnersMode(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "owners", Title: "Owners"},
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.tabIdx != tabQuote || m.quoteCenterMode != quoteCenterOwners {
		t.Fatalf("expected Quote owners mode, got tab=%d mode=%d", m.tabIdx, m.quoteCenterMode)
	}
	if cmd == nil {
		t.Fatal("expected owners load command")
	}
}

func TestQuoteOwnersViewRendersTopHoldersAndPreview(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 140
	model.height = 40
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterOwners
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.quote.Symbol = "AAPL"
	model.quote.Price = 210
	model.owners = sampleOwnershipSnapshot()

	view := model.View()
	if !strings.Contains(view, "OWNERS") {
		t.Fatal("expected owners section")
	}
	if !strings.Contains(view, "TOP HOLDERS") || !strings.Contains(view, "Breakdown") {
		t.Fatal("expected holders list and ownership summary")
	}
	if !strings.Contains(view, "Vanguard Group") || !strings.Contains(view, "Selected holder") {
		t.Fatal("expected selected owner preview")
	}
	if strings.Contains(view, "SYMBOLS") || strings.Contains(view, "\nNEWS\n") {
		t.Fatal("expected owners page to hide quote sidebars")
	}
}

func TestQuoteOwnersModeUsesArrowNavigationInsteadOfWatchlist(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabQuote
	model.quoteCenterMode = quoteCenterOwners
	model.config.Watchlist = []string{"AAPL", "MSFT"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	model.owners = sampleOwnershipSnapshot()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(Model)
	if m.selectedIdx != 0 {
		t.Fatalf("expected watchlist selection unchanged, got %d", m.selectedIdx)
	}
	if m.ownersSel != 1 {
		t.Fatalf("expected owners selection to move, got %d", m.ownersSel)
	}
}
