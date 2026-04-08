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

func TestSearchModeClearsInputOnOpenAndEsc(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})

	model.searchInput.SetValue("stale")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m := updated.(Model)
	if !m.searchMode {
		t.Fatal("expected search mode to be enabled")
	}
	if m.searchInput.Value() != "" {
		t.Fatalf("expected cleared search input, got %q", m.searchInput.Value())
	}

	m.searchInput.SetValue("aapl")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.searchMode {
		t.Fatal("expected search mode to be disabled")
	}
	if m.searchInput.Value() != "" {
		t.Fatalf("expected cleared search input after esc, got %q", m.searchInput.Value())
	}
}

func TestViewShowsSearchInStatusBar(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 120
	model.height = 40
	model.searchMode = true
	model.searchInput.Focus()
	model.searchInput.SetValue("aapl")
	model.status = "Loaded"

	view := model.View()
	if !strings.Contains(view, "aapl") {
		t.Fatal("expected search input to be rendered")
	}
	if strings.Contains(view, "Keys: / search") {
		t.Fatal("expected key help to be hidden while searching")
	}
	if strings.Contains(view, "Loaded") {
		t.Fatal("expected status line to be hidden while searching")
	}
}

func TestSearchModeAllowsTypingJK(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})
	model.searchMode = true
	model.searchInput.Focus()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m := updated.(Model)
	if m.searchInput.Value() != "k" {
		t.Fatalf("expected typed k in search input, got %q", m.searchInput.Value())
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.searchInput.Value() != "kj" {
		t.Fatalf("expected typed j in search input, got %q", m.searchInput.Value())
	}
}

func TestSearchModeClearsPreviousResultsWhenQueryChanges(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config: storage.DefaultConfig(),
	})
	model.searchMode = true
	model.searchInput.Focus()
	model.searchItems = []domain.SymbolRef{{Symbol: "AAPL"}, {Symbol: "AAP"}}
	model.searchIdx = 1

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m := updated.(Model)
	if len(m.searchItems) != 0 {
		t.Fatalf("expected stale search results to clear, got %d", len(m.searchItems))
	}
	if m.searchIdx != 0 {
		t.Fatalf("expected search index reset, got %d", m.searchIdx)
	}
	if m.searchInput.Value() != "x" {
		t.Fatalf("expected updated query, got %q", m.searchInput.Value())
	}
}

func TestViewShowsSearchResultsInBottomPanel(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.width = 120
	model.height = 40
	model.searchMode = true
	model.searchInput.Focus()
	model.searchInput.SetValue("aapl")
	model.searchItems = []domain.SymbolRef{
		{Symbol: "AAPL", Name: "Apple Inc.", Exchange: "NMS", Type: "EQUITY"},
	}

	view := model.View()
	if !strings.Contains(view, "SEARCH RESULTS") {
		t.Fatal("expected dedicated search results panel")
	}
	if strings.Contains(view, "\nRESULTS\n") {
		t.Fatal("expected sidebar results section to be removed")
	}
	if !strings.Contains(view, "Apple Inc.") {
		t.Fatal("expected search result details to be rendered")
	}
}
