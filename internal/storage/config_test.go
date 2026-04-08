package storage

import (
	"fmt"
	"os"
	"testing"
)

func TestConfigStoreLoadSave(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("HOME", root)

	store, err := NewConfigStore("blackdesk-test")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ActiveSymbol == "" {
		t.Fatal("expected default active symbol")
	}
	if len(cfg.Watchlist) < 20 {
		t.Fatalf("expected fuller default watchlist, got %d symbols", len(cfg.Watchlist))
	}

	cfg.ActiveSymbol = "TSLA"
	cfg.Watchlist = []string{"TSLA", "AAPL"}
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(store.file)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected saved config bytes")
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.ActiveSymbol != "TSLA" {
		t.Fatalf("expected TSLA, got %s", reloaded.ActiveSymbol)
	}
}

func TestConfigStoreLoadMigratesLegacyDefaultWatchlist(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("HOME", root)

	store, err := NewConfigStore("blackdesk-test")
	if err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Watchlist = append([]string(nil), legacyDefaultWatchlist...)
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Watchlist) <= len(legacyDefaultWatchlist) {
		t.Fatalf("expected migrated fuller watchlist, got %d symbols", len(reloaded.Watchlist))
	}
}

func TestConfigStoreLoadTrimsWatchlistToMaxAndRepairsActiveSymbol(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("HOME", root)

	store, err := NewConfigStore("blackdesk-test")
	if err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Watchlist = nil
	for i := 0; i < MaxWatchlistItems+5; i++ {
		cfg.Watchlist = append(cfg.Watchlist, fmt.Sprintf("SYM%02d", i))
	}
	cfg.ActiveSymbol = cfg.Watchlist[len(cfg.Watchlist)-1]
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(reloaded.Watchlist); got != MaxWatchlistItems {
		t.Fatalf("expected watchlist length %d, got %d", MaxWatchlistItems, got)
	}
	if reloaded.ActiveSymbol != reloaded.Watchlist[0] {
		t.Fatalf("expected active symbol repaired to %s, got %s", reloaded.Watchlist[0], reloaded.ActiveSymbol)
	}
}

func TestConfigStoreResetRestoresDefaults(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	t.Setenv("HOME", root)

	store, err := NewConfigStore("blackdesk-test")
	if err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.ActiveSymbol = "TSLA"
	cfg.Watchlist = []string{"TSLA", "AAPL"}
	if err := store.Save(cfg); err != nil {
		t.Fatal(err)
	}

	reset, err := store.Reset()
	if err != nil {
		t.Fatal(err)
	}
	if reset.ActiveSymbol != DefaultConfig().ActiveSymbol {
		t.Fatalf("expected default active symbol %s, got %s", DefaultConfig().ActiveSymbol, reset.ActiveSymbol)
	}
	if len(reset.Watchlist) != len(DefaultConfig().Watchlist) {
		t.Fatalf("expected default watchlist length %d, got %d", len(DefaultConfig().Watchlist), len(reset.Watchlist))
	}

	reloaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.ActiveSymbol != DefaultConfig().ActiveSymbol {
		t.Fatalf("expected persisted default active symbol %s, got %s", DefaultConfig().ActiveSymbol, reloaded.ActiveSymbol)
	}
	if reloaded.AIConnector != "" {
		t.Fatalf("expected reset AI connector to be empty, got %q", reloaded.AIConnector)
	}
}

func TestDefaultConfigStartsWithoutAISelection(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.AIConnector != "" {
		t.Fatalf("expected empty default AI connector, got %q", cfg.AIConnector)
	}
	if cfg.AIModel != "" {
		t.Fatalf("expected empty default AI model, got %q", cfg.AIModel)
	}
}
