package application

import (
	"strings"
	"testing"

	"blackdesk/internal/storage"
)

func TestAddWatchlistSymbolEvictsOldestWhenLimitExceeded(t *testing.T) {
	cfg := storage.DefaultConfig()
	cfg.Watchlist = nil
	for i := 0; i < storage.MaxWatchlistItems; i++ {
		cfg.Watchlist = append(cfg.Watchlist, strings.ToUpper(string(rune('A'+(i%26))))+strings.Repeat("X", 3))
	}
	oldest := cfg.Watchlist[len(cfg.Watchlist)-1]

	state := AddWatchlistSymbol(cfg, 0, 0, 10, "TSLA")

	if got := len(state.Config.Watchlist); got != storage.MaxWatchlistItems {
		t.Fatalf("expected watchlist capped at %d, got %d", storage.MaxWatchlistItems, got)
	}
	if state.Config.Watchlist[0] != "TSLA" {
		t.Fatalf("expected newest symbol at front, got %s", state.Config.Watchlist[0])
	}
	for _, item := range state.Config.Watchlist {
		if item == oldest {
			t.Fatalf("expected oldest symbol %s to be evicted", oldest)
		}
	}
}

func TestAddWatchlistSymbolSelectsExistingAndScrollsIntoView(t *testing.T) {
	cfg := storage.DefaultConfig()

	state := AddWatchlistSymbol(cfg, 0, 0, 5, "TSLA")

	if state.SelectedIndex != 10 {
		t.Fatalf("expected existing TSLA index 10, got %d", state.SelectedIndex)
	}
	if state.Config.ActiveSymbol != "TSLA" {
		t.Fatalf("expected active symbol TSLA, got %s", state.Config.ActiveSymbol)
	}
	if state.Scroll == 0 {
		t.Fatal("expected scroll to move to existing selected symbol")
	}
}

func TestRemoveWatchlistSymbolRepairsSelectionAndActiveSymbol(t *testing.T) {
	cfg := storage.Config{
		Watchlist:    []string{"AAPL", "MSFT", "NVDA"},
		ActiveSymbol: "NVDA",
	}

	state := RemoveWatchlistSymbol(cfg, 2, 1, 2)

	if got := strings.Join(state.Config.Watchlist, ","); got != "AAPL,MSFT" {
		t.Fatalf("unexpected watchlist after removal: %s", got)
	}
	if state.Config.ActiveSymbol != "MSFT" {
		t.Fatalf("expected active symbol repaired to MSFT, got %s", state.Config.ActiveSymbol)
	}
	if state.SelectedIndex != 1 {
		t.Fatalf("expected selected index 1, got %d", state.SelectedIndex)
	}
}

func TestSupplementalMarketQuoteSymbolsDedupesWatchlistAndActiveSymbol(t *testing.T) {
	got := SupplementalMarketQuoteSymbols(
		[]string{"AAPL", "MSFT"},
		"SPY",
		[]string{"SPY", "QQQ", "AAPL", "DIA", "QQQ"},
	)

	want := []string{"QQQ", "DIA"}
	if len(got) != len(want) {
		t.Fatalf("expected %d symbols, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected symbol %d to be %s, got %s", i, want[i], got[i])
		}
	}
}
