package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"blackdesk/internal/domain"
)

type Config struct {
	DefaultProvider string   `json:"default_provider"`
	AIConnector     string   `json:"ai_connector"`
	AIModel         string   `json:"ai_model"`
	ActiveSymbol    string   `json:"active_symbol"`
	Watchlist       []string `json:"watchlist"`
	RefreshSeconds  int      `json:"refresh_seconds"`
	DefaultRange    string   `json:"default_range"`
	DefaultInterval string   `json:"default_interval"`
}

type ConfigStore struct {
	dir  string
	file string
}

const MaxWatchlistItems = 25

var legacyDefaultWatchlist = []string{"AAPL", "MSFT", "NVDA", "SPY"}

func NewConfigStore(appName string) (*ConfigStore, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(root, appName)
	return &ConfigStore{
		dir:  dir,
		file: filepath.Join(dir, "config.json"),
	}, nil
}

func DefaultConfig() Config {
	return Config{
		DefaultProvider: "yahoo",
		AIConnector:     "",
		AIModel:         "",
		ActiveSymbol:    "SPY",
		Watchlist: []string{
			"SPY", "QQQ", "DIA", "IWM", "AAPL", "MSFT", "NVDA", "AMZN", "GOOGL", "META",
			"TSLA", "AVGO", "BRK-B", "JPM", "LLY", "V", "MA", "NFLX", "XOM", "COST",
			"WMT", "ORCL", "HD", "UNH", "AMD",
		},
		RefreshSeconds:  domain.DefaultRefreshSeconds,
		DefaultRange:    "1mo",
		DefaultInterval: "1d",
	}
}

func (s *ConfigStore) Load() (Config, error) {
	if _, err := os.Stat(s.file); errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig()
		return cfg, s.Save(cfg)
	}

	data, err := os.ReadFile(s.file)
	if err != nil {
		return Config{}, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.RefreshSeconds <= 0 {
		cfg.RefreshSeconds = domain.DefaultRefreshSeconds
	}
	if len(cfg.Watchlist) == 0 {
		cfg.Watchlist = DefaultConfig().Watchlist
	} else if isLegacyDefaultWatchlist(cfg.Watchlist) {
		cfg.Watchlist = DefaultConfig().Watchlist
	}
	cfg.Watchlist = trimWatchlist(cfg.Watchlist)
	if cfg.ActiveSymbol == "" {
		cfg.ActiveSymbol = cfg.Watchlist[0]
	}
	if !containsSymbol(cfg.Watchlist, cfg.ActiveSymbol) {
		cfg.ActiveSymbol = cfg.Watchlist[0]
	}
	if cfg.DefaultRange == "" {
		cfg.DefaultRange = "1mo"
	}
	if cfg.DefaultInterval == "" {
		cfg.DefaultInterval = "1d"
	}
	return cfg, nil
}

func (s *ConfigStore) Save(cfg Config) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0o644)
}

func isLegacyDefaultWatchlist(items []string) bool {
	if len(items) != len(legacyDefaultWatchlist) {
		return false
	}
	for i, item := range items {
		if item != legacyDefaultWatchlist[i] {
			return false
		}
	}
	return true
}

func trimWatchlist(items []string) []string {
	if len(items) <= MaxWatchlistItems {
		return items
	}
	return append([]string(nil), items[:MaxWatchlistItems]...)
}

func containsSymbol(items []string, symbol string) bool {
	for _, item := range items {
		if item == symbol {
			return true
		}
	}
	return false
}
