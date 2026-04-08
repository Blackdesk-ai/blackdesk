package tui

import (
	"encoding/json"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

func (m Model) buildAIQuoteInsightRequest(symbol string) (RequestEnvelope, error) {
	fullGuide := aiContextGuide()
	statements := m.statementBundleForSymbol(symbol)
	statementInsights := aiStatementInsights(statements)
	payloadStruct := struct {
		GeneratedAt       string                      `json:"generated_at"`
		MarketProvider    string                      `json:"market_provider"`
		AIConnector       string                      `json:"ai_connector"`
		AIModel           string                      `json:"ai_model"`
		ActiveTab         string                      `json:"active_tab"`
		ActiveSymbol      string                      `json:"active_symbol"`
		SelectedRange     string                      `json:"selected_range"`
		SelectedInterval  string                      `json:"selected_interval"`
		ActiveQuote       domain.QuoteSnapshot        `json:"active_quote"`
		Fundamentals      domain.FundamentalsSnapshot `json:"fundamentals"`
		Statements        aiStatementsSnapshot        `json:"statements"`
		StatementInsights []aiStatRow                 `json:"statement_insights"`
		QuoteStats        map[string]string           `json:"quote_stats"`
		TechnicalValues   map[string]string           `json:"technical_values"`
		Technicals        map[string][]aiStatRow      `json:"technicals"`
		TechnicalLookup   map[string]aiStatRow        `json:"technical_lookup"`
		MarketNews        []domain.NewsItem           `json:"market_news"`
		News              []domain.NewsItem           `json:"news"`
		ContextGuide      map[string]string           `json:"context_guide"`
		StatRowGuide      map[string]string           `json:"stat_row_guide"`
	}{
		GeneratedAt:       time.Now().Format(time.RFC3339),
		MarketProvider:    m.statusMetaMarketSource(),
		AIConnector:       m.activeAIConnectorID(),
		AIModel:           strings.TrimSpace(m.config.AIModel),
		ActiveTab:         "quote and research",
		ActiveSymbol:      symbol,
		SelectedRange:     ranges[m.rangeIdx].Range,
		SelectedInterval:  ranges[m.rangeIdx].Interval,
		ActiveQuote:       m.quote,
		Fundamentals:      m.fundamentals,
		Statements:        statements,
		StatementInsights: statementInsights,
		QuoteStats:        m.aiQuoteStats(),
		TechnicalValues:   m.aiTechnicalValues(),
		Technicals:        m.aiTechnicalSections(),
		TechnicalLookup:   m.aiTechnicalLookup(),
		MarketNews:        compactAINews(append([]domain.NewsItem(nil), m.marketNews...), 8),
		News:              compactAINews(append([]domain.NewsItem(nil), m.news...), 10),
		ContextGuide: map[string]string{
			"generated_at":       fullGuide["generated_at"],
			"market_provider":    fullGuide["market_provider"],
			"ai_connector":       fullGuide["ai_connector"],
			"ai_model":           fullGuide["ai_model"],
			"active_tab":         fullGuide["active_tab"],
			"active_symbol":      fullGuide["active_symbol"],
			"selected_range":     fullGuide["selected_range"],
			"selected_interval":  fullGuide["selected_interval"],
			"active_quote":       "Normalized quote snapshot for the active symbol. Use this with fundamentals and technicals for the rating call.",
			"fundamentals":       "Full normalized fundamentals snapshot for the active symbol, including fields that may not currently be visible in the Quote sidebar UI, such as ROIC and invested capital when available.",
			"statements":         "Normalized financial statement bundle for the active symbol. Use annual and quarterly income, balance-sheet, and cash-flow sections for revenue, margins, leverage, liquidity, and cash generation confirmation.",
			"statement_insights": "Derived high-signal statement metrics computed from the statement bundle. Use these first for quick read-through of growth, cash generation, leverage, and liquidity.",
			"quote_stats":        fullGuide["quote_stats"],
			"technicals":         "Technical indicator sections for the active symbol. Use these for trend, momentum, volatility, and volume confirmation.",
			"technical_lookup":   fullGuide["technical_lookup"],
			"technical_values":   fullGuide["technical_values"],
			"market_news":        fullGuide["market_news"],
			"news":               fullGuide["news"],
		},
		StatRowGuide: aiStatRowGuide(),
	}
	ctxPayload, err := json.MarshalIndent(payloadStruct, "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}
	payload := truncateRunes(string(ctxPayload), aiMaxContextChars)

	return RequestEnvelope{
		Prompt:         "Write a short AI quote insight for the active symbol using the provided company context.",
		SystemPrompt:   truncateRunes(buildAIQuoteInsightSystemPrompt(payload), aiMaxPromptChars),
		ContextPayload: payload,
		ActiveSymbol:   symbol,
	}, nil
}

func buildAIQuoteInsightSystemPrompt(payload string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(aiSystemPromptTemplate))
	b.WriteString("\n\n")
	b.WriteString("This request is for the Quote sidebar AI Insight widget.\n")
	b.WriteString("Use only the provided single-company Blackdesk context for the active symbol.\n")
	b.WriteString("Start the response with exactly one stance label followed by a colon: Buy:, Hold:, Reduce:, Sell:, or Watchlist:.\n")
	b.WriteString("The stance must be explicit and easy to scan.\n")
	b.WriteString("Base the call on the combined evidence from price action, technicals, fundamentals, financial statements, valuation, analyst target context, and recent company news.\n")
	b.WriteString("Use `statement_insights` first for the fast read, then verify with the full fundamentals object and statement bundle when needed.\n")
	b.WriteString("If the signals conflict, state the main conflict briefly after the stance.\n")
	b.WriteString("Return exactly one very short paragraph of one or two tight sentences and about 140 to 170 characters.\n")
	b.WriteString("Prefer fewer words over completeness.\n")
	b.WriteString("No bullets, no headers, no sources, no markdown table, no caveat dump.\n\n")
	b.WriteString("<blackdesk_context_update>\n")
	b.WriteString(payload)
	b.WriteString("\n</blackdesk_context_update>\n")
	return b.String()
}
