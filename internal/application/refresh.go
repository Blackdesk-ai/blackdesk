package application

import "time"

type AutoRefreshInput struct {
	Now                   time.Time
	LastAutoRefresh       time.Time
	LastMarketNewsRefresh time.Time
	RefreshSeconds        int
	DefaultRefreshSeconds int
	NewsTabActive         bool
	ScreenerTabActive     bool
	ScreenerLoaded        bool
	MarketNewsInterval    time.Duration
}

type AutoRefreshPlan struct {
	RefreshAll            bool
	RefreshScreener       bool
	RefreshMarketNews     bool
	NextLastAutoRefresh   time.Time
	NextLastMarketRefresh time.Time
}

func PlanAutoRefresh(input AutoRefreshInput) AutoRefreshPlan {
	refreshEvery := time.Duration(input.RefreshSeconds) * time.Second
	if refreshEvery <= 0 {
		refreshEvery = time.Duration(input.DefaultRefreshSeconds) * time.Second
	}

	plan := AutoRefreshPlan{
		NextLastAutoRefresh:   input.LastAutoRefresh,
		NextLastMarketRefresh: input.LastMarketNewsRefresh,
	}

	if input.Now.Sub(input.LastAutoRefresh) >= refreshEvery {
		plan.RefreshAll = true
		plan.NextLastAutoRefresh = input.Now
		if input.ScreenerTabActive && input.ScreenerLoaded {
			plan.RefreshScreener = true
		}
	}

	if input.NewsTabActive && input.Now.Sub(input.LastMarketNewsRefresh) >= input.MarketNewsInterval {
		plan.RefreshMarketNews = true
		plan.NextLastMarketRefresh = input.Now
	}

	return plan
}
