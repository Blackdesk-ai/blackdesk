package application

type WatchlistNavigationInput struct {
	CurrentIndex int
	Count        int
	Step         int
}

type WatchlistNavigationResult struct {
	NextIndex int
	Changed   bool
}

func PlanWatchlistNavigation(input WatchlistNavigationInput) WatchlistNavigationResult {
	result := WatchlistNavigationResult{NextIndex: input.CurrentIndex}
	if input.Count <= 0 || input.Step == 0 {
		return result
	}
	next := input.CurrentIndex + input.Step
	if next < 0 || next >= input.Count {
		return result
	}
	result.NextIndex = next
	result.Changed = true
	return result
}

type WrappedIndexInput struct {
	CurrentIndex int
	Count        int
	Step         int
}

type WrappedIndexResult struct {
	NextIndex int
	Changed   bool
}

func PlanWrappedIndexStep(input WrappedIndexInput) WrappedIndexResult {
	result := WrappedIndexResult{NextIndex: input.CurrentIndex}
	if input.Count <= 0 || input.Step == 0 {
		return result
	}
	result.NextIndex = (input.CurrentIndex + input.Step + input.Count) % input.Count
	result.Changed = result.NextIndex != input.CurrentIndex
	return result
}
