package application

type ScreenerSymbolAction string

const (
	ScreenerSymbolOpenQuote ScreenerSymbolAction = "open_quote"
	ScreenerSymbolAddWatch  ScreenerSymbolAction = "add_watchlist"
)

type ScreenerSymbolActionInput struct {
	Action  ScreenerSymbolAction
	HasItem bool
	Symbol  string
}

type ScreenerSymbolActionResult struct {
	ApplyStatus  bool
	Status       string
	SelectSymbol bool
	OpenQuote    bool
	AddWatchlist bool
}

func PlanScreenerSymbolAction(input ScreenerSymbolActionInput) ScreenerSymbolActionResult {
	if !input.HasItem || input.Symbol == "" {
		return ScreenerSymbolActionResult{}
	}
	switch input.Action {
	case ScreenerSymbolAddWatch:
		return ScreenerSymbolActionResult{
			ApplyStatus:  true,
			Status:       "Added " + input.Symbol + " to watchlist",
			AddWatchlist: true,
		}
	default:
		return ScreenerSymbolActionResult{
			ApplyStatus:  true,
			Status:       "Opened " + input.Symbol + " in Quote",
			AddWatchlist: true,
			SelectSymbol: true,
			OpenQuote:    true,
		}
	}
}
