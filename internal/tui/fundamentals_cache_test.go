package tui

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestLoadFundamentalsCmdUsesCachedSnapshot(t *testing.T) {
	provider := &aiPrepProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	snapshot := domain.FundamentalsSnapshot{Symbol: "AAPL", Description: "Cached fundamentals", MarketCap: 1}
	model.cacheFundamentals(snapshot)

	cmd := model.loadFundamentalsCmd("AAPL")
	if cmd == nil {
		t.Fatal("expected fundamentals command")
	}
	msg, ok := cmd().(fundamentalsLoadedMsg)
	if !ok {
		t.Fatalf("expected fundamentalsLoadedMsg, got %T", cmd())
	}
	if provider.fundamentalsCalls != 0 {
		t.Fatalf("expected cached fundamentals to skip provider call, got %d", provider.fundamentalsCalls)
	}
	if msg.data.Description != "Cached fundamentals" || msg.err != nil {
		t.Fatalf("unexpected cached fundamentals message %+v", msg)
	}
}

func TestHandleFundamentalsLoadedKeepsCachedSnapshotOnError(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(testProvider{}),
	})
	model.config.Watchlist = []string{"AAPL"}
	model.config.ActiveSymbol = "AAPL"
	model.selectedIdx = 0
	snapshot := domain.FundamentalsSnapshot{Symbol: "AAPL", Description: "Last good data", MarketCap: 1}
	model.fundamentals = snapshot
	model.cacheFundamentals(snapshot)

	updated, _ := model.handleFundamentalsLoaded(fundamentalsLoadedMsg{
		symbol: "AAPL",
		err:    errors.New("timeout"),
	})

	if updated.fundamentals.Description != "Last good data" {
		t.Fatalf("expected cached fundamentals to survive transient error, got %+v", updated.fundamentals)
	}
	if updated.errFundamentals == nil {
		t.Fatal("expected fundamentals error to still be recorded")
	}
}

func TestTickAutoRefreshOnlyQueuesQuoteLoads(t *testing.T) {
	provider := &aiPrepProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	now := time.Now()
	oldMarketNews := now.Add(-20 * time.Minute)
	model.lastAutoRefresh = now.Add(-2 * time.Minute)
	model.lastMarketNews = oldMarketNews

	updated, cmd := model.Update(tickMsg(now))
	model = updated.(Model)
	if cmd == nil {
		t.Fatal("expected auto-refresh command")
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected batched auto-refresh command, got %T", msg)
	}
	for _, subcmd := range batch {
		if subcmd == nil {
			continue
		}
		submsg := subcmd()
		inner, ok := submsg.(tea.BatchMsg)
		if !ok {
			continue
		}
		for _, innerCmd := range inner {
			if innerCmd == nil {
				continue
			}
			_ = innerCmd()
		}
	}
	if provider.quoteCalls != 1 {
		t.Fatalf("expected one active quote refresh, got %d", provider.quoteCalls)
	}
	if provider.quotesCalls != 2 {
		t.Fatalf("expected watchlist and market quote refreshes, got %d", provider.quotesCalls)
	}
	if provider.newsCalls != 0 || provider.fundamentalsCalls != 0 {
		t.Fatalf("expected auto-refresh to skip news and fundamentals, got news=%d fundamentals=%d", provider.newsCalls, provider.fundamentalsCalls)
	}
	if !model.lastMarketNews.Equal(oldMarketNews) {
		t.Fatalf("expected market news refresh clock to stay unchanged, got %v want %v", model.lastMarketNews, oldMarketNews)
	}
}
