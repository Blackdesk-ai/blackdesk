package application

type QuoteCenterMode string

const (
	QuoteCenterChart        QuoteCenterMode = "chart"
	QuoteCenterFundamentals QuoteCenterMode = "fundamentals"
	QuoteCenterTechnicals   QuoteCenterMode = "technicals"
	QuoteCenterStatistics   QuoteCenterMode = "statistics"
	QuoteCenterStatements   QuoteCenterMode = "statements"
	QuoteCenterInsiders     QuoteCenterMode = "insiders"
	QuoteCenterOwners       QuoteCenterMode = "owners"
	QuoteCenterAnalyst      QuoteCenterMode = "analyst"
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
	LoadOwners          bool
	LoadAnalyst         bool
	LoadFilings         bool
	LoadEarnings        bool
}

func PlanQuoteWorkspaceLoad(mode QuoteCenterMode, needsTechnical, hasStatements, hasInsiders, hasOwners bool) QuoteWorkspaceLoadPlan {
	plan := QuoteWorkspaceLoadPlan{
		LoadQuote:           true,
		LoadWatchlistQuotes: true,
		LoadMarketQuotes:    true,
		LoadHistory:         true,
		LoadNews:            true,
		LoadFundamentals:    true,
	}
	if (mode == QuoteCenterTechnicals || mode == QuoteCenterFundamentals) && needsTechnical {
		plan.LoadTechnical = true
	}
	if mode == QuoteCenterStatements && hasStatements {
		plan.LoadStatement = true
	}
	if mode == QuoteCenterInsiders && hasInsiders {
		plan.LoadInsiders = true
	}
	if mode == QuoteCenterOwners && hasOwners {
		plan.LoadOwners = true
	}
	if mode == QuoteCenterAnalyst {
		plan.LoadAnalyst = true
	}
	if mode == QuoteCenterFilings {
		plan.LoadFilings = true
	}
	if mode == QuoteCenterEarnings {
		plan.LoadEarnings = true
	}
	return plan
}
