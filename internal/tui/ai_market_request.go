package tui

import (
	"encoding/json"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

func (m Model) buildAIMarketOpinionRequest(histories map[string]domain.PriceSeries) (RequestEnvelope, error) {
	fullGuide := aiContextGuide()
	payloadStruct := struct {
		GeneratedAt         string                 `json:"generated_at"`
		MarketProvider      string                 `json:"market_provider"`
		AIConnector         string                 `json:"ai_connector"`
		AIModel             string                 `json:"ai_model"`
		ActiveTab           string                 `json:"active_tab"`
		Markets             map[string][]aiStatRow `json:"markets"`
		MarketNews          []domain.NewsItem      `json:"market_news"`
		MarketRegimeSignals map[string]string      `json:"market_regime_signals"`
		ContextGuide        map[string]string      `json:"context_guide"`
		StatRowGuide        map[string]string      `json:"stat_row_guide"`
	}{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		MarketProvider: m.statusMetaMarketSource(),
		AIConnector:    m.activeAIConnectorID(),
		AIModel:        strings.TrimSpace(m.config.AIModel),
		ActiveTab:      "global market board",
		Markets:        m.aiMarketSections(),
		MarketNews:     compactAINews(append([]domain.NewsItem(nil), m.marketNews...), 8),
		MarketRegimeSignals: marketRegimeSignals(
			m.watchQuotes,
			histories,
		),
		ContextGuide: map[string]string{
			"generated_at":          fullGuide["generated_at"],
			"market_provider":       fullGuide["market_provider"],
			"ai_connector":          fullGuide["ai_connector"],
			"ai_model":              fullGuide["ai_model"],
			"active_tab":            fullGuide["active_tab"],
			"markets":               fullGuide["markets"],
			"market_news":           fullGuide["market_news"],
			"market_regime_signals": "Derived multi-session regime cues for the markets widget, including trend-vs-SMA200 context and selected cross-asset direction checks.",
		},
		StatRowGuide: aiStatRowGuide(),
	}
	ctxPayload, err := json.MarshalIndent(payloadStruct, "", "  ")
	if err != nil {
		return RequestEnvelope{}, err
	}
	payload := truncateRunes(string(ctxPayload), aiMaxContextChars)

	return RequestEnvelope{
		Prompt:         "Write a very short AI insight for the Markets sidebar using only the provided market board context.",
		SystemPrompt:   truncateRunes(buildAIMarketOpinionSystemPrompt(payload), aiMaxPromptChars),
		ContextPayload: payload,
		ActiveSymbol:   "",
	}, nil
}

func buildAIMarketOpinionSystemPrompt(payload string) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(aiSystemPromptTemplate))
	b.WriteString("\n\n")
	b.WriteString("This request is for the Markets sidebar AI Insight widget.\n")
	b.WriteString("Use only the provided market board context.\n")
	b.WriteString("Use both `markets` and `market_regime_signals`, and take `market_regime_signals` seriously as higher-level regime context derived from the basket history.\n")
	b.WriteString("Start the response with exactly one regime label followed by a colon: Risk-on:, Risk-off:, or Mixed:.\n")
	b.WriteString("The regime label must be explicit and easy to scan.\n")
	b.WriteString("Return exactly one short paragraph of one or two sentences and about 160 characters.\n")
	b.WriteString("Add value through interpretation, not paraphrase.\n")
	b.WriteString("Do not label the regime or restate the panel mechanically.\n")
	b.WriteString("Focus on what is non-obvious, what signal matters most, what contradiction stands out, or what to watch next.\n")
	b.WriteString("Prioritize the strongest signals and any important contradiction across volatility, breadth, futures, yields, credit, commodities, FX, regions, and sectors.\n")
	b.WriteString("No bullets, no headers, no sources, no caveat dump.\n\n")
	b.WriteString("<blackdesk_context_update>\n")
	b.WriteString(payload)
	b.WriteString("\n</blackdesk_context_update>\n")
	return b.String()
}
