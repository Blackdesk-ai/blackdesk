package application

import "testing"

func TestPlanScreenerSymbolActionBuildsOpenAndAddIntents(t *testing.T) {
	open := PlanScreenerSymbolAction(ScreenerSymbolActionInput{
		Action:  ScreenerSymbolOpenQuote,
		HasItem: true,
		Symbol:  "AAPL",
	})
	if !open.OpenQuote || !open.SelectSymbol || !open.AddWatchlist || open.Status != "Opened AAPL in Quote" {
		t.Fatalf("unexpected open action plan: %+v", open)
	}

	add := PlanScreenerSymbolAction(ScreenerSymbolActionInput{
		Action:  ScreenerSymbolAddWatch,
		HasItem: true,
		Symbol:  "MSFT",
	})
	if add.OpenQuote || !add.AddWatchlist || add.Status != "Added MSFT to watchlist" {
		t.Fatalf("unexpected add action plan: %+v", add)
	}
}
