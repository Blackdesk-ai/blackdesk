package application

import "testing"

func TestPlanWatchlistNavigationBoundsAndMovement(t *testing.T) {
	moved := PlanWatchlistNavigation(WatchlistNavigationInput{CurrentIndex: 1, Count: 4, Step: 1})
	if !moved.Changed || moved.NextIndex != 2 {
		t.Fatalf("unexpected forward watchlist move: %+v", moved)
	}

	blocked := PlanWatchlistNavigation(WatchlistNavigationInput{CurrentIndex: 0, Count: 4, Step: -1})
	if blocked.Changed || blocked.NextIndex != 0 {
		t.Fatalf("unexpected out-of-bounds watchlist move: %+v", blocked)
	}
}

func TestPlanWrappedIndexStepWrapsAround(t *testing.T) {
	result := PlanWrappedIndexStep(WrappedIndexInput{CurrentIndex: 0, Count: 5, Step: -1})
	if !result.Changed || result.NextIndex != 4 {
		t.Fatalf("unexpected wrapped step result: %+v", result)
	}
}
