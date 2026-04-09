package tui

import (
	"blackdesk/internal/agents"
	"blackdesk/internal/application"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

type Dependencies struct {
	Services           *application.Services
	MarketRiskProvider application.MarketRiskProvider
	Registry           *providers.Registry
	AgentRegistry      *agents.Registry
	ConfigStore        *storage.ConfigStore
	Config             storage.Config
	WorkspaceRoot      string
}
