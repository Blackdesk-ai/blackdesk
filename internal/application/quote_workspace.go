package application

type QuoteCenterMode string

const (
	QuoteCenterChart        QuoteCenterMode = "chart"
	QuoteCenterFundamentals QuoteCenterMode = "fundamentals"
	QuoteCenterTechnicals   QuoteCenterMode = "technicals"
	QuoteCenterStatements   QuoteCenterMode = "statements"
	QuoteCenterInsiders     QuoteCenterMode = "insiders"
	QuoteCenterFilings      QuoteCenterMode = "filings"
	QuoteCenterEarnings     QuoteCenterMode = "earnings"
	QuoteCenterNews         QuoteCenterMode = "news"
)

type QuoteWorkspaceLoadPlan struct {
	LoadQuote           bool
	LoadWatchlistQuotes bool
	LoadMarketQuotes    bool
	LoadHistory         bool
	LoadNews            bool
	LoadFundamentals    bool
	LoadTechnical       bool
	LoadStatement       bool
	LoadInsiders        bool
	LoadFilings         bool
	LoadEarnings        bool
}

func PlanQuoteWorkspaceLoad(mode QuoteCenterMode, needsTechnical, hasStatements, hasInsiders bool) QuoteWorkspaceLoadPlan {
	plan := QuoteWorkspaceLoadPlan{
		LoadQuote:           true,
		LoadWatchlistQuotes: true,
		LoadMarketQuotes:    true,
		LoadHistory:         true,
		LoadNews:            true,
		LoadFundamentals:    true,
	}
	if mode == QuoteCenterTechnicals && needsTechnical {
		plan.LoadTechnical = true
	}
	if mode == QuoteCenterStatements && hasStatements {
		plan.LoadStatement = true
	}
	if mode == QuoteCenterInsiders && hasInsiders {
		plan.LoadInsiders = true
	}
	if mode == QuoteCenterFilings {
		plan.LoadFilings = true
	}
	if mode == QuoteCenterEarnings {
		plan.LoadEarnings = true
	}
	return plan
}
