package application

import (
	"testing"
	"time"
)

func TestPlanAutoRefreshQueuesAllRelevantWorkspaceRefreshes(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	plan := PlanAutoRefresh(AutoRefreshInput{
		Now:                   now,
		LastAutoRefresh:       now.Add(-2 * time.Minute),
		LastMarketNewsRefresh: now.Add(-16 * time.Minute),
		RefreshSeconds:        60,
		DefaultRefreshSeconds: 60,
		NewsTabActive:         true,
		ScreenerTabActive:     true,
		ScreenerLoaded:        true,
		MarketNewsInterval:    15 * time.Minute,
	})

	if !plan.RefreshAll || !plan.RefreshScreener || !plan.RefreshMarketNews {
		t.Fatalf("unexpected refresh plan: %+v", plan)
	}
	if !plan.NextLastAutoRefresh.Equal(now) || !plan.NextLastMarketRefresh.Equal(now) {
		t.Fatalf("expected refresh clocks to advance to now, got %+v", plan)
	}
}

func TestPlanAutoRefreshSkipsInactiveWorkspaceRefreshes(t *testing.T) {
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	plan := PlanAutoRefresh(AutoRefreshInput{
		Now:                   now,
		LastAutoRefresh:       now.Add(-2 * time.Minute),
		LastMarketNewsRefresh: now.Add(-16 * time.Minute),
		RefreshSeconds:        60,
		DefaultRefreshSeconds: 60,
		NewsTabActive:         false,
		ScreenerTabActive:     false,
		ScreenerLoaded:        true,
		MarketNewsInterval:    15 * time.Minute,
	})

	if !plan.RefreshAll {
		t.Fatal("expected main refresh to run")
	}
	if plan.RefreshScreener || plan.RefreshMarketNews {
		t.Fatalf("expected workspace refreshes to stay disabled, got %+v", plan)
	}
}
