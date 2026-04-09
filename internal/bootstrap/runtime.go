package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"blackdesk/internal/agents"
	"blackdesk/internal/application"
	"blackdesk/internal/providers"
	"blackdesk/internal/providers/blackdeskapi"
	"blackdesk/internal/providers/composite"
	"blackdesk/internal/providers/rss"
	"blackdesk/internal/providers/yahoo"
	"blackdesk/internal/storage"
	"blackdesk/internal/tui"
)

const appName = "blackdesk"

func LoadDependencies(ctx context.Context) (tui.Dependencies, error) {
	cfgStore, err := storage.NewConfigStore(appName)
	if err != nil {
		return tui.Dependencies{}, err
	}

	cfg, err := cfgStore.Load()
	if err != nil {
		return tui.Dependencies{}, err
	}

	yahooProvider := yahoo.New(yahoo.Config{
		BaseURL: "https://query1.finance.yahoo.com",
		Client:  nil,
		Cache:   storage.NewMemoryCache(),
		Timeout: 10 * time.Second,
	})
	provider := composite.New(
		yahooProvider,
		rss.New(rss.Config{
			Cache:   storage.NewMemoryCache(),
			Timeout: 10 * time.Second,
		}),
	)

	workspaceRoot, err := os.Getwd()
	if err != nil {
		return tui.Dependencies{}, err
	}

	return tui.Dependencies{
		Services: application.NewServices(providers.NewRegistry(provider), agents.NewRegistry(), cfgStore),
		MarketRiskProvider: blackdeskapi.NewRiskProvider(blackdeskapi.RiskConfig{
			Cache:   storage.NewMemoryCache(),
			Timeout: 4 * time.Second,
		}),
		Config:        cfg,
		WorkspaceRoot: filepath.Clean(workspaceRoot),
	}, nil
}
