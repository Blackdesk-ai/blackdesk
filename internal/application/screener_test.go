package application

import (
	"testing"

	"blackdesk/internal/domain"
)

func TestPlanScreenerEntryHandlesInitialLoadAndReuse(t *testing.T) {
	load := PlanScreenerEntry(ScreenerEntryInput{Available: true, WasActive: false, HasItems: false})
	if !load.MarkLoaded || !load.ShouldLoad || load.Status != "Loading screener…" {
		t.Fatalf("unexpected initial screener entry plan: %+v", load)
	}

	reuse := PlanScreenerEntry(ScreenerEntryInput{Available: true, WasActive: false, HasItems: true, CurrentName: "Most Active"})
	if reuse.ShouldLoad || reuse.Status != "Screener: Most Active" {
		t.Fatalf("unexpected screener reuse plan: %+v", reuse)
	}
}

func TestPlanScreenerAdvanceRotatesAndBuildsLoadStatus(t *testing.T) {
	result := PlanScreenerAdvance(ScreenerAdvanceInput{
		Available: true,
		Definitions: []domain.ScreenerDefinition{
			{ID: "most_actives", Name: "Most Active"},
			{ID: "day_gainers", Name: "Day Gainers"},
		},
		CurrentIndex: 0,
		Step:         1,
	})

	if result.NextIndex != 1 || !result.ShouldLoad || result.Status != "Loading screener: Day Gainers" {
		t.Fatalf("unexpected screener advance result: %+v", result)
	}
}

func TestApplyScreenerLoadBuildsStatusAndRepairsSelection(t *testing.T) {
	result := ApplyScreenerLoad(ScreenerLoadInput{
		SelectedIndex: 4,
		UserTriggered: true,
		Data: domain.ScreenerResult{
			Definition: domain.ScreenerDefinition{Name: "Most Active"},
			Items:      []domain.ScreenerItem{{Symbol: "AAPL"}},
		},
	})

	if !result.Loaded || result.SelectedIndex != 0 || result.Status != "Loaded Most Active (1 names)" {
		t.Fatalf("unexpected screener load result: %+v", result)
	}
}
