package tui

import (
	"strings"
	"time"

	"blackdesk/internal/domain"
)

func (m Model) aiContextSnapshot() aiSnapshot {
	statements := m.statementBundleForSymbol(m.activeSymbol())
	statementInsights := aiStatementInsights(statements)
	watchlist := append([]string(nil), m.config.Watchlist...)
	watchQuotes := make(map[string]domain.QuoteSnapshot, len(m.watchQuotes))
	for symbol, quote := range m.watchQuotes {
		watchQuotes[symbol] = quote
	}
	news := compactAINews(append([]domain.NewsItem(nil), m.news...), aiContextNewsItems)
	marketNews := compactAINews(append([]domain.NewsItem(nil), m.marketNews...), aiContextMarketNewsItems)
	searchResults := append([]domain.SymbolRef(nil), m.searchItems...)
	if len(searchResults) > aiContextSearchResults {
		searchResults = searchResults[:aiContextSearchResults]
	}
	return aiSnapshot{
		GeneratedAt:       time.Now().Format(time.RFC3339),
		ContextGuide:      aiContextGuide(),
		StatRowGuide:      aiStatRowGuide(),
		MarketProvider:    m.statusMetaMarketSource(),
		MarketRegime:      aiMarketRegimeFromSnapshot(m.marketRisk),
		AIConnector:       m.activeAIConnectorID(),
		AIModel:           strings.TrimSpace(m.config.AIModel),
		Status:            m.status,
		ActiveTab:         tabIdentity(m.tabIdx),
		ActiveSymbol:      m.activeSymbol(),
		SelectedRange:     ranges[m.rangeIdx].Range,
		SelectedInterval:  ranges[m.rangeIdx].Interval,
		Watchlist:         watchlist,
		WatchQuotes:       watchQuotes,
		ActiveQuote:       m.quote,
		Fundamentals:      m.fundamentals,
		Statements:        statements,
		Insiders:          m.insidersForSymbol(m.activeSymbol()),
		StatementInsights: statementInsights,
		QuoteStats:        m.aiQuoteStats(),
		Technicals:        m.aiTechnicalSections(),
		TechnicalLookup:   m.aiTechnicalLookup(),
		TechnicalValues:   m.aiTechnicalValues(),
		Markets:           m.aiMarketSections(),
		MarketNews:        marketNews,
		News:              news,
		SearchResults:     searchResults,
	}
}

func aiContextGuide() map[string]string {
	return map[string]string{
		"generated_at":       "RFC3339 timestamp for when the Blackdesk snapshot was generated.",
		"market_provider":    "Market data source currently backing the app, such as Yahoo.",
		"market_regime":      "External market regime snapshot from Blackdesk risk service. `score` runs from -4 to +4, where 0 is neutral, values above 0 are risk-on, values below 0 are risk-off, and larger absolute values mean a stronger regime. `components` shows each signed contribution to the total score, `inputs` shows the raw market series used to score it, and `thresholds` shows the decision buffers used by the model.",
		"ai_connector":       "Selected local AI connector id, such as codex, claude, or opencode.",
		"ai_model":           "Selected model name passed to the local AI connector.",
		"status":             "Current Blackdesk status line text at request time.",
		"active_tab":         "Current app workspace or tab name.",
		"active_symbol":      "Primary symbol currently selected in Blackdesk.",
		"selected_range":     "Current historical range key for the active symbol chart, such as 1mo or 6mo.",
		"selected_interval":  "Current candle interval key for the active symbol chart, such as 1d or 1wk.",
		"watchlist":          "Ordered list of user watchlist symbols.",
		"watch_quotes":       "Latest quote snapshots keyed by symbol for items Blackdesk has loaded.",
		"active_quote":       "Normalized quote snapshot for the active symbol.",
		"fundamentals":       "Normalized fundamentals snapshot for the active symbol, including valuation, margins, liquidity, leverage, and return metrics such as ROA, ROE, ROIC, and invested capital when available.",
		"statements":         "Normalized financial statement bundle for the active symbol with income, balance sheet, and cash flow data across annual and quarterly views when available.",
		"insiders":           "Normalized insider activity snapshot for the active symbol, including recent insider transactions with explicit action labels such as Buy or Sale when derivable from provider text, roster holders, and six-month purchase or sale summary when the provider supports it.",
		"statement_insights": "Derived high-signal statement metrics for the active symbol, such as revenue growth, earnings trend, cash flow trend, leverage, and liquidity, computed from the normalized statement bundle.",
		"quote_stats":        "Compact quote stat labels rendered in the quote view for the active symbol. This also includes derived fields such as `Earnings Yield` and `Growth Est.` when the needed inputs are available. `Earnings Yield` is the inverse of trailing PE when present, with a fallback to EPS divided by price. `Growth Est.` is the market-implied 5-year EPS CAGR band reverse-engineered from forward PE, trailing PE, or both, and it represents expectations already embedded in the current price rather than a Blackdesk forecast.",
		"technicals":         "Technical indicator sections for the active symbol, grouped by momentum, trend, stat_edge, volatility, and volume.",
		"technical_lookup":   "Normalized technical indicator lookup keyed by compact aliases such as HV21, HV63, ATR14, RSI14, PRICEZ21, HVRANK, and HVPCTL.",
		"technical_values":   "Direct alias-to-value map for active symbol technical indicators. Use this first when the user asks for the exact value of an indicator such as HV21, ATR14, RSI14, ADX14, or PRICEZ21.",
		"markets":            "Global market board sections from the Markets tab, grouped by futures, yields, credit, commodities, fx, volatility, regions, countries, and sectors.",
		"market_news":        "Recent market-wide headlines from the dedicated News workspace, aggregated from normalized RSS/Atom sources and sorted newest first.",
		"news":               "Recent news items for the active symbol.",
		"search_results":     "Recent symbol search results shown in Blackdesk.",
	}
}
