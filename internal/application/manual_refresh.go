package application

type Workspace string

const (
	WorkspaceQuote    Workspace = "quote"
	WorkspaceMarkets  Workspace = "markets"
	WorkspaceNews     Workspace = "news"
	WorkspaceScreener Workspace = "screener"
)

type ManualRefreshInput struct {
	Workspace         Workspace
	ActiveSymbol      string
	ScreenerAvailable bool
}

type ManualRefreshPlan struct {
	ApplyStatus       bool
	Status            string
	RefreshAll        bool
	RefreshScreener   bool
	RefreshMarketNews bool
	RefreshMarketSnap bool
	TouchNewsClock    bool
}

func PlanManualRefresh(input ManualRefreshInput) ManualRefreshPlan {
	switch input.Workspace {
	case WorkspaceScreener:
		if !input.ScreenerAvailable {
			return ManualRefreshPlan{
				ApplyStatus: true,
				Status:      "Screeners unavailable for active provider",
			}
		}
		return ManualRefreshPlan{
			ApplyStatus:     true,
			Status:          "Refreshing screener…",
			RefreshScreener: true,
		}
	case WorkspaceNews:
		return ManualRefreshPlan{
			ApplyStatus:       true,
			Status:            "Refreshing market news…",
			RefreshMarketNews: true,
			RefreshMarketSnap: true,
			TouchNewsClock:    true,
		}
	case WorkspaceMarkets:
		return ManualRefreshPlan{
			ApplyStatus: true,
			Status:      "Refreshing market data…",
			RefreshAll:  true,
		}
	default:
		return ManualRefreshPlan{
			ApplyStatus: true,
			Status:      "Refreshing " + input.ActiveSymbol + "…",
			RefreshAll:  true,
		}
	}
}
