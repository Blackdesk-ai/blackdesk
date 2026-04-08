package application

import "testing"

func TestPlanManualRefreshBuildsWorkspaceSpecificActions(t *testing.T) {
	news := PlanManualRefresh(ManualRefreshInput{Workspace: WorkspaceNews, ActiveSymbol: "AAPL"})
	if !news.RefreshMarketNews || !news.RefreshMarketSnap || !news.TouchNewsClock || news.Status != "Refreshing market news…" {
		t.Fatalf("unexpected news refresh plan: %+v", news)
	}

	markets := PlanManualRefresh(ManualRefreshInput{Workspace: WorkspaceMarkets, ActiveSymbol: "SPY"})
	if !markets.RefreshAll || markets.Status != "Refreshing market data…" {
		t.Fatalf("unexpected markets refresh plan: %+v", markets)
	}

	quote := PlanManualRefresh(ManualRefreshInput{Workspace: WorkspaceQuote, ActiveSymbol: "NVDA"})
	if !quote.RefreshAll || quote.Status != "Refreshing NVDA…" {
		t.Fatalf("unexpected quote refresh plan: %+v", quote)
	}
}

func TestPlanManualRefreshRejectsUnavailableScreener(t *testing.T) {
	plan := PlanManualRefresh(ManualRefreshInput{Workspace: WorkspaceScreener, ScreenerAvailable: false})
	if plan.RefreshScreener || plan.Status != "Screeners unavailable for active provider" {
		t.Fatalf("unexpected screener refresh plan: %+v", plan)
	}
}
