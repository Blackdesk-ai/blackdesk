package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestCtrlKOpensCommandPalette(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m := updated.(Model)
	if !m.commandPaletteOpen {
		t.Fatal("expected command palette to open")
	}
	if len(m.commandPaletteItems) == 0 {
		t.Fatal("expected command palette to preload function items")
	}
}

func TestCommandPaletteEnterOpensSelectedFunction(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "fundamentals", Title: "Fundamentals"},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.commandPaletteOpen {
		t.Fatal("expected command palette to close after opening a function")
	}
	if m.tabIdx != tabQuote {
		t.Fatalf("expected Quote tab, got %d", m.tabIdx)
	}
	if m.quoteCenterMode != quoteCenterFundamentals {
		t.Fatalf("expected fundamentals mode, got %d", m.quoteCenterMode)
	}
}

func TestCommandPaletteEnterOpensSelectedSymbol(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemSymbol, Symbol: domain.SymbolRef{Symbol: "AAPL", Name: "Apple Inc."}, Title: "AAPL"},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(Model)
	if m.commandPaletteOpen {
		t.Fatal("expected command palette to close after opening a symbol")
	}
	if m.config.ActiveSymbol != "AAPL" {
		t.Fatalf("expected active symbol AAPL, got %q", m.config.ActiveSymbol)
	}
	if m.tabIdx != tabQuote {
		t.Fatalf("expected Quote tab, got %d", m.tabIdx)
	}
}

func TestCommandPaletteTypingSchedulesDebouncedSymbolSearch(t *testing.T) {
	provider := &searchProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m := updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)

	if len(provider.queries) != 0 {
		t.Fatalf("expected no immediate symbol search, got %d queries", len(provider.queries))
	}

	updated, cmd := m.Update(commandPaletteDebouncedMsg{id: m.commandPaletteDebounceID, query: "a"})
	m = updated.(Model)
	if cmd == nil {
		t.Fatal("expected debounced command palette search command")
	}
	msg := cmd()
	loaded, ok := msg.(commandPaletteLoadedMsg)
	if !ok {
		t.Fatalf("expected commandPaletteLoadedMsg, got %T", msg)
	}
	if loaded.query != "a" {
		t.Fatalf("expected query a, got %q", loaded.query)
	}
	if len(provider.queries) != 1 || provider.queries[0] != "a" {
		t.Fatalf("expected symbol search for a, got %#v", provider.queries)
	}
}

func TestCommandPaletteViewRendersFullscreenPage(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})
	model.width = 120
	model.height = 40
	model.commandPaletteOpen = true
	model.commandInput.Focus()
	model.commandPaletteItems = []commandPaletteItem{
		{Kind: commandPaletteItemFunction, FunctionID: "quote", Title: "Quote", Meta: "Function • Workspace", Subtitle: "Open the Quote workspace"},
	}

	view := model.View()
	if !strings.Contains(view, "COMMAND PALETTE") {
		t.Fatal("expected command palette title in view")
	}
	if !strings.Contains(view, "RESULTS") || !strings.Contains(view, "PREVIEW") {
		t.Fatal("expected results and preview sections in command palette view")
	}
}

func TestCommandPaletteIncludesFilingsWhenAvailable(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:          storage.DefaultConfig(),
		FilingsProvider: filingsProvider{},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m := updated.(Model)
	found := false
	for _, item := range m.commandPaletteItems {
		if item.Kind == commandPaletteItemFunction && item.FunctionID == "filings" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected filings function in command palette")
	}
}

func TestCommandPaletteIncludesAnalystRecommendationsWhenAvailable(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m := updated.(Model)
	found := false
	for _, item := range m.commandPaletteItems {
		if item.Kind == commandPaletteItemFunction && item.FunctionID == "analyst" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected analyst recommendations function in command palette")
	}
}

func TestCommandPaletteIncludesCalendarWhenAvailable(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlK})
	m := updated.(Model)
	found := false
	for _, item := range m.commandPaletteItems {
		if item.Kind == commandPaletteItemFunction && item.FunctionID == "calendar" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected calendar function in command palette")
	}
}
