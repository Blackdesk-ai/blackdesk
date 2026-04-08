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

func TestNumberKeyThreeSelectsNewsTab(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabMarkets

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(Model)

	if model.tabIdx != tabNews {
		t.Fatalf("expected key 3 to select News tab, got %d", model.tabIdx)
	}
}

func TestInitDoesNotLoadMarketNewsOutsideNewsTab(t *testing.T) {
	provider := &marketNewsProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("expected init command")
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected init to return batched commands, got %T", msg)
	}
	if len(batch) != 3 {
		t.Fatalf("expected init to queue quote load, ticker, and version check outside News tab, got %d commands", len(batch))
	}
}

func TestEnteringNewsTabLoadsMarketNewsAndMarketSnap(t *testing.T) {
	provider := &marketNewsProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	before := model.lastMarketNews

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(Model)

	if model.tabIdx != tabNews {
		t.Fatalf("expected key 3 to select News tab, got %d", model.tabIdx)
	}
	if !model.lastMarketNews.After(before) {
		t.Fatal("expected entering News tab to reset market news refresh clock")
	}
	if cmd == nil {
		t.Fatal("expected entering News tab to queue market data loads")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected entering News tab to batch market loads, got %T", msg)
	}
	for _, subcmd := range batch {
		if subcmd == nil {
			continue
		}
		_ = subcmd()
	}
	if provider.marketNewsCalls != 1 {
		t.Fatalf("expected one market news request when opening News tab, got %d", provider.marketNewsCalls)
	}
	if provider.quotesCalls != 1 {
		t.Fatalf("expected one market snapshot quote request when opening News tab, got %d", provider.quotesCalls)
	}
}

func TestNewsTabShowsMarketWireKeyHint(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.tabIdx = tabNews
	model.width = 140
	model.height = 40
	model.marketNews = []domain.NewsItem{
		{Title: "Futures rise ahead of CPI", Publisher: "Desk", Time: time.Now()},
	}

	view := ansi.Strip(model.View())
	if !strings.Contains(view, "MARKET WIRE ↑/↓") {
		t.Fatal("expected market wire key hint in news tab header")
	}
}

func TestTickDoesNotAutoRefreshMarketNewsOutsideNewsTab(t *testing.T) {
	provider := &marketNewsProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.tabIdx = tabMarkets
	now := time.Now()
	staleAt := now.Add(-marketNewsRefreshInterval - time.Second)
	model.lastAutoRefresh = now
	model.lastMarketNews = staleAt

	updated, _ := model.Update(tickMsg(now))
	model = updated.(Model)

	if !model.lastMarketNews.Equal(staleAt) {
		t.Fatal("expected market news refresh clock to stay unchanged outside News tab")
	}
}

func TestFilterMarketNewsRecentKeepsLast24HoursAndZeroTimestamp(t *testing.T) {
	now := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	items := []domain.NewsItem{
		{Title: "Recent", Time: time.Date(2026, 4, 6, 9, 15, 0, 0, time.UTC)},
		{Title: "Within 24h", Time: time.Date(2026, 4, 5, 20, 30, 0, 0, time.UTC)},
		{Title: "Old", Time: time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)},
		{Title: "No timestamp"},
	}

	got := filterMarketNewsRecent(items, now)
	if len(got) != 3 {
		t.Fatalf("expected 3 items (2 recent + 1 no timestamp), got %d", len(got))
	}
	if got[0].Title != "Recent" {
		t.Fatalf("expected Recent first, got %+v", got[0])
	}
	if got[1].Title != "Within 24h" {
		t.Fatalf("expected Within 24h second, got %+v", got[1])
	}
	if got[2].Title != "No timestamp" {
		t.Fatalf("expected No timestamp third, got %+v", got[2])
	}
}
