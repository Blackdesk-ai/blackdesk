package tui

import (
	_ "embed"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

const (
	aiMaxPromptChars          = 1000000
	aiMaxContextChars         = 220000
	aiMaxHistoryChars         = 700000
	aiMaxMessageChars         = 64000
	aiMaxSummaryChars         = 180000
	aiMaxRecentHistoryChars   = 520000
	aiRecentHistoryMinMsgs    = 8
	aiSummaryUserChars        = 1200
	aiSummaryAssistantChars   = 1800
	aiFilingPromptChars       = 700000
	aiFilingDocumentChars     = 550000
	aiContextNewsItems        = 20
	aiContextMarketNewsItems  = 24
	aiContextSearchResults    = 20
	aiQuoteInsightNewsItems   = 20
	aiQuoteInsightMarketItems = 16
	aiMarketOpinionNewsItems  = 16
	aiNewsTitleChars          = 320
	aiNewsSummaryChars        = 560
	aiNewsPublisherChars      = 96
	aiNewsURLChars            = 320
)

//go:embed prompts/ai_system.md
var aiSystemPromptTemplate string

type aiSnapshot struct {
	GeneratedAt       string                          `json:"generated_at"`
	MarketProvider    string                          `json:"market_provider"`
	MarketRegime      aiMarketRegime                  `json:"market_regime"`
	AIConnector       string                          `json:"ai_connector"`
	AIModel           string                          `json:"ai_model"`
	Status            string                          `json:"status"`
	ActiveTab         string                          `json:"active_tab"`
	ActiveSymbol      string                          `json:"active_symbol"`
	SelectedRange     string                          `json:"selected_range"`
	SelectedInterval  string                          `json:"selected_interval"`
	ActiveQuote       domain.QuoteSnapshot            `json:"active_quote"`
	Fundamentals      domain.FundamentalsSnapshot     `json:"fundamentals"`
	Statements        aiStatementsSnapshot            `json:"statements"`
	Insiders          domain.InsiderSnapshot          `json:"insiders"`
	StatementInsights []aiStatRow                     `json:"statement_insights"`
	QuoteStats        map[string]string               `json:"quote_stats"`
	TechnicalValues   map[string]string               `json:"technical_values"`
	Technicals        map[string][]aiStatRow          `json:"technicals"`
	TechnicalLookup   map[string]aiStatRow            `json:"technical_lookup"`
	Markets           map[string][]aiStatRow          `json:"markets"`
	MarketNews        []domain.NewsItem               `json:"market_news"`
	Watchlist         []string                        `json:"watchlist"`
	WatchQuotes       map[string]domain.QuoteSnapshot `json:"watch_quotes"`
	News              []domain.NewsItem               `json:"news"`
	SearchResults     []domain.SymbolRef              `json:"search_results"`
	ContextGuide      map[string]string               `json:"context_guide"`
	StatRowGuide      map[string]string               `json:"stat_row_guide"`
}

type aiMarketRegime struct {
	Available      bool                         `json:"available"`
	Score          int                          `json:"score"`
	Min            int                          `json:"min"`
	Max            int                          `json:"max"`
	Stance         string                       `json:"stance"`
	Strength       string                       `json:"strength"`
	Interpretation string                       `json:"interpretation"`
	SourceLabel    string                       `json:"source_label,omitempty"`
	Thresholds     aiMarketRiskThresholds       `json:"thresholds"`
	Components     map[string]int               `json:"components,omitempty"`
	Inputs         map[string]aiMarketRiskInput `json:"inputs,omitempty"`
	GeneratedAt    string                       `json:"generated_at,omitempty"`
	MarketNow      string                       `json:"market_now,omitempty"`
	MarketTimezone string                       `json:"market_timezone,omitempty"`
	MarketCalendar string                       `json:"market_calendar,omitempty"`
}

type aiMarketRiskThresholds struct {
	SMABufferPct    float64 `json:"sma_buffer_pct"`
	Breadth50Buffer float64 `json:"breadth_50_buffer"`
}

type aiMarketRiskInput struct {
	Name    string  `json:"name"`
	Symbol  string  `json:"symbol"`
	Current float64 `json:"current"`
	SMA200  float64 `json:"sma200"`
}

type aiStatementsSnapshot struct {
	IncomeAnnual          domain.FinancialStatement `json:"income_annual"`
	BalanceSheetAnnual    domain.FinancialStatement `json:"balance_sheet_annual"`
	CashFlowAnnual        domain.FinancialStatement `json:"cash_flow_annual"`
	IncomeQuarterly       domain.FinancialStatement `json:"income_quarterly"`
	BalanceSheetQuarterly domain.FinancialStatement `json:"balance_sheet_quarterly"`
	CashFlowQuarterly     domain.FinancialStatement `json:"cash_flow_quarterly"`
}

type aiStatRow struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Signal string `json:"signal,omitempty"`
}

type aiMessageRole string

const (
	aiMessageUser      aiMessageRole = "user"
	aiMessageAssistant aiMessageRole = "assistant"
)

type aiMessage struct {
	Role      aiMessageRole
	Body      string
	Timestamp time.Time
	Meta      string
}

func (s aiStatementsSnapshot) loadedCount() int {
	count := 0
	for _, stmt := range []domain.FinancialStatement{
		s.IncomeAnnual,
		s.BalanceSheetAnnual,
		s.CashFlowAnnual,
		s.IncomeQuarterly,
		s.BalanceSheetQuarterly,
		s.CashFlowQuarterly,
	} {
		if strings.TrimSpace(stmt.Symbol) != "" && len(stmt.Rows) > 0 {
			count++
		}
	}
	return count
}
