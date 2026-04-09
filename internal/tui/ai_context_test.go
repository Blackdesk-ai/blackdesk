package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"blackdesk/internal/domain"
	"blackdesk/internal/providers"
	"blackdesk/internal/storage"
)

func TestAIContextSnapshotIncludesMarkets(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.watchQuotes["SPY"] = domain.QuoteSnapshot{Symbol: "SPY", Price: 500, ChangePercent: 1.2}
	model.marketRisk = domain.MarketRiskSnapshot{
		Score:      2,
		Min:        -4,
		Max:        4,
		Available:  true,
		Components: map[string]int{"spy_vs_sma200": 1},
		Inputs: map[string]domain.MarketRiskInput{
			"spy": {Name: "SPY", Symbol: "SPY", Current: 500, SMA200: 480},
		},
	}

	snapshot := model.aiContextSnapshot()
	if len(snapshot.Markets) == 0 {
		t.Fatal("expected market sections in AI context snapshot")
	}
	if _, ok := snapshot.Markets["futures"]; !ok {
		t.Fatal("expected futures market section in AI context snapshot")
	}
	if snapshot.MarketRegime.Stance != "risk_on" {
		t.Fatalf("expected external market regime in AI context, got %+v", snapshot.MarketRegime)
	}
	if snapshot.MarketRegime.Components["spy_vs_sma200"] != 1 || snapshot.MarketRegime.Inputs["spy"].Symbol != "SPY" {
		t.Fatalf("expected market regime internals in AI context, got %+v", snapshot.MarketRegime)
	}
}

func TestAIContextSnapshotNeverIncludesScreenerData(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.screenerLoaded = true
	model.screenerResult = domain.ScreenerResult{
		Definition: domain.ScreenerDefinition{
			ID:          "screener_only_fixture_id",
			Name:        "Screener Only Fixture",
			Category:    "Market Movers",
			Description: "Most traded names",
			Kind:        "equity",
		},
		Total: 42,
		Items: []domain.ScreenerItem{
			{
				Symbol:        "ZZZSCREEN",
				Name:          "Screener Fixture Corp",
				Price:         177.64,
				ChangePercent: 1.2,
				Metrics:       []domain.ScreenerMetric{{Key: "avg_3m_volume", Label: "Avg 3M Vol", Value: "181.1M"}},
			},
		},
		Criteria: []domain.ScreenerCriterion{
			{Statement: "Market cap >= $2B"},
		},
		UpdatedAt: time.Now(),
	}

	snapshot := model.aiContextSnapshot()
	if _, ok := snapshot.ContextGuide["screener"]; ok {
		t.Fatal("expected screener guide entry to stay out of AI context")
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot failed: %v", err)
	}
	if strings.Contains(string(payload), "\"screener\"") || strings.Contains(string(payload), "screener_only_fixture_id") || strings.Contains(string(payload), "ZZZSCREEN") {
		t.Fatal("expected screener fields and values to stay out of AI context snapshot")
	}
}

func TestBuildAIRequestExcludesScreenerDataFromPayload(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.screenerLoaded = true
	model.screenerResult = domain.ScreenerResult{
		Definition: domain.ScreenerDefinition{
			ID:   "screener_only_fixture_id",
			Name: "Screener Only Fixture",
		},
		Items: []domain.ScreenerItem{{Symbol: "ZZZSCREEN", Name: "Screener Fixture Corp"}},
	}

	req, err := model.buildAIRequest("summarize the setup")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if strings.Contains(req.ContextPayload, "\"screener\"") || strings.Contains(req.ContextPayload, "screener_only_fixture_id") || strings.Contains(req.ContextPayload, "ZZZSCREEN") {
		t.Fatal("expected AI request payload to exclude screener data")
	}
}

func TestBuildAIMarketOpinionRequestUsesMarketOnlyContext(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.watchQuotes["SPY"] = domain.QuoteSnapshot{Symbol: "SPY", Price: 500, ChangePercent: 1.2}
	model.marketRisk = domain.MarketRiskSnapshot{
		Score:      -1,
		Min:        -4,
		Max:        4,
		Available:  true,
		Thresholds: domain.MarketRiskThresholds{SMABufferPct: 1, Breadth50Buffer: 2},
		Components: map[string]int{"s5th_vs_sma200": -1},
		Inputs: map[string]domain.MarketRiskInput{
			"s5th": {Name: "S&P 500 Stocks Above 200-Day Average", Symbol: "$S5TH", Current: 54.98, SMA200: 59},
		},
	}

	req, err := model.buildAIMarketOpinionRequest(nil)
	if err != nil {
		t.Fatalf("buildAIMarketOpinionRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "Markets sidebar AI Insight widget") {
		t.Fatal("expected dedicated market insight instructions")
	}
	if !strings.Contains(req.ContextPayload, "\"markets\"") {
		t.Fatal("expected market opinion payload to include markets section")
	}
	if !strings.Contains(req.ContextPayload, "\"market_regime\"") {
		t.Fatal("expected market opinion payload to include external market regime")
	}
	if !strings.Contains(req.ContextPayload, "\"components\"") || !strings.Contains(req.ContextPayload, "\"inputs\"") || !strings.Contains(req.ContextPayload, "\"thresholds\"") {
		t.Fatal("expected market opinion payload to include full regime breakdown")
	}
	if strings.Contains(req.ContextPayload, "\"market_regime_signals\"") {
		t.Fatal("expected old tape-style regime signals to be removed from market opinion payload")
	}
	if strings.Contains(req.ContextPayload, "\"active_quote\"") || strings.Contains(req.ContextPayload, "\"fundamentals\"") || strings.Contains(req.ContextPayload, "\"news\"") {
		t.Fatal("expected market opinion payload to stay market-only")
	}
}

func TestBuildAIQuoteInsightRequestUsesFullCompanyContext(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.quote = domain.QuoteSnapshot{Symbol: "AAPL", Price: 210, TrailingPEGRatio: 2.1}
	model.fundamentals = domain.FundamentalsSnapshot{
		Symbol:            "AAPL",
		Sector:            "Technology",
		Industry:          "Consumer Electronics",
		GrossMargins:      0.46,
		ProfitMargins:     0.26,
		OperatingMargins:  0.31,
		RecommendationKey: "buy",
		TargetMeanPrice:   225,
	}
	model.news = []domain.NewsItem{{Title: "Apple expands AI push"}}
	model.marketRisk = domain.MarketRiskSnapshot{
		Score:      -2,
		Min:        -4,
		Max:        4,
		Available:  true,
		Thresholds: domain.MarketRiskThresholds{SMABufferPct: 1, Breadth50Buffer: 2},
		Components: map[string]int{"spy_vs_sma200": 1, "hyg_vs_sma200": 0},
		Inputs: map[string]domain.MarketRiskInput{
			"spy": {Name: "SPY", Symbol: "SPY", Current: 210, SMA200: 205},
		},
	}
	populateAllStatementCache(&model, "AAPL")
	candles := make([]domain.Candle, 0, 260)
	for i := 260; i >= 0; i-- {
		price := 100.0 + float64(260-i)*0.5
		candles = append(candles, domain.Candle{
			Time:   time.Now().AddDate(0, 0, -i),
			Open:   price - 1,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1_000_000,
		})
	}
	model.series = domain.PriceSeries{Symbol: "AAPL", Range: "3mo", Interval: "1d", Candles: candles}
	model.technicalCache["AAPL"] = domain.PriceSeries{Symbol: "AAPL", Range: "2y", Interval: "1d", Candles: candles}

	req, err := model.buildAIQuoteInsightRequest("AAPL")
	if err != nil {
		t.Fatalf("buildAIQuoteInsightRequest failed: %v", err)
	}
	if !strings.Contains(req.SystemPrompt, "Quote sidebar AI Insight widget") {
		t.Fatal("expected dedicated quote insight instructions")
	}
	if !strings.Contains(req.SystemPrompt, "Buy:, Hold:, Reduce:, Sell:, or Watchlist:") {
		t.Fatal("expected explicit stance instructions")
	}
	if !strings.Contains(req.ContextPayload, "\"fundamentals\"") || !strings.Contains(req.ContextPayload, "\"technicals\"") || !strings.Contains(req.ContextPayload, "\"statements\"") {
		t.Fatal("expected quote insight payload to include fundamentals, statements, and technicals")
	}
	if !strings.Contains(req.ContextPayload, "\"market_regime\"") {
		t.Fatal("expected quote insight payload to include market regime context")
	}
	if !strings.Contains(req.SystemPrompt, "market_regime") {
		t.Fatal("expected quote insight prompt to explicitly instruct use of market regime")
	}
	if !strings.Contains(req.ContextPayload, "\"Industry\": \"Consumer Electronics\"") {
		t.Fatal("expected hidden fundamentals fields to remain in quote insight context")
	}
	if !strings.Contains(req.ContextPayload, "\"GrossMargins\": 0.46") {
		t.Fatal("expected margin fields to remain in quote insight context")
	}
	if !strings.Contains(req.ContextPayload, "\"Label\": \"Total Revenue\"") {
		t.Fatal("expected financial statement rows to remain in quote insight context")
	}
	if !strings.Contains(req.ContextPayload, "\"income_annual\"") || !strings.Contains(req.ContextPayload, "\"cash_flow_quarterly\"") {
		t.Fatal("expected quote insight payload to include the full financial statement bundle")
	}
	if !strings.Contains(req.ContextPayload, "\"statement_insights\"") || !strings.Contains(req.ContextPayload, "\"Revenue YoY\"") {
		t.Fatal("expected quote insight payload to include derived statement insights")
	}
}

func TestAIContextSnapshotIncludesTechnicalLookupAliases(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.quote = domain.QuoteSnapshot{Symbol: "AAPL", Price: 210}
	model.series = domain.PriceSeries{Symbol: "AAPL"}
	for i := 260; i >= 0; i-- {
		price := 100.0 + float64(260-i)*0.5
		model.series.Candles = append(model.series.Candles, domain.Candle{
			Time:   time.Now().AddDate(0, 0, -i),
			Open:   price - 1,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1_000_000,
		})
	}

	snapshot := model.aiContextSnapshot()
	row, ok := snapshot.TechnicalLookup["HV21"]
	if !ok {
		t.Fatal("expected normalized HV21 alias in technical lookup")
	}
	if row.Name != "HV 21" {
		t.Fatalf("expected HV21 alias to resolve to HV 21 row, got %q", row.Name)
	}
	if snapshot.TechnicalValues["HV21"] == "" {
		t.Fatal("expected direct HV21 value in technical_values")
	}
}

func TestAIContextSnapshotIncludesStatements(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	populateAllStatementCache(&model, model.activeSymbol())

	snapshot := model.aiContextSnapshot()
	if snapshot.Statements.IncomeAnnual.Symbol != model.activeSymbol() {
		t.Fatal("expected annual income statement to be included in AI context snapshot")
	}
	if snapshot.Statements.CashFlowQuarterly.Frequency != domain.StatementFrequencyQuarterly {
		t.Fatal("expected quarterly cash flow statement to remain normalized in AI context snapshot")
	}
	if len(snapshot.StatementInsights) == 0 {
		t.Fatal("expected derived statement insights in AI context snapshot")
	}
	if snapshot.StatementInsights[0].Name == "" || snapshot.StatementInsights[0].Value == "" {
		t.Fatal("expected derived statement insights to keep name/value pairs")
	}
}

func TestAIContextSnapshotIncludesInsiders(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.cacheInsiders(sampleInsiderSnapshot(model.activeSymbol()))

	snapshot := model.aiContextSnapshot()
	if snapshot.Insiders.Symbol != model.activeSymbol() {
		t.Fatal("expected insider snapshot to be included in AI context")
	}
	if len(snapshot.Insiders.Transactions) == 0 || len(snapshot.Insiders.Roster) == 0 {
		t.Fatal("expected insider transactions and roster in AI context")
	}
	if snapshot.Insiders.Transactions[0].Action == "" {
		t.Fatal("expected insider transactions in AI context to expose normalized action labels")
	}
}

func TestAIContextSnapshotIncludesGuides(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})

	snapshot := model.aiContextSnapshot()
	if snapshot.ContextGuide["insiders"] == "" {
		t.Fatal("expected context guide entry for insiders")
	}
	if snapshot.ContextGuide["statements"] == "" {
		t.Fatal("expected context guide entry for statements")
	}
	if snapshot.ContextGuide["statement_insights"] == "" {
		t.Fatal("expected context guide entry for statement_insights")
	}
	if snapshot.ContextGuide["technical_lookup"] == "" {
		t.Fatal("expected context guide entry for technical_lookup")
	}
	if snapshot.ContextGuide["markets"] == "" {
		t.Fatal("expected context guide entry for markets")
	}
	if snapshot.StatRowGuide["value"] == "" {
		t.Fatal("expected stat row guide entry for value")
	}
}

func TestAIRequestPrioritizesTechnicalValuesInContextPayload(t *testing.T) {
	model := NewModel(context.Background(), Dependencies{Config: storage.DefaultConfig()})
	model.quote = domain.QuoteSnapshot{Symbol: "AAPL", Price: 210}
	model.series = domain.PriceSeries{Symbol: "AAPL"}
	for i := 260; i >= 0; i-- {
		price := 100.0 + float64(260-i)*0.5
		model.series.Candles = append(model.series.Candles, domain.Candle{
			Time:   time.Now().AddDate(0, 0, -i),
			Open:   price - 1,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1_000_000,
		})
	}
	for i := 0; i < 400; i++ {
		symbol := fmt.Sprintf("SYM%03d", i)
		model.watchQuotes[symbol] = domain.QuoteSnapshot{
			Symbol:    symbol,
			ShortName: strings.Repeat("X", 40),
			Price:     100 + float64(i),
		}
	}

	req, err := model.buildAIRequest("what is hv21")
	if err != nil {
		t.Fatalf("buildAIRequest failed: %v", err)
	}
	if !strings.Contains(req.ContextPayload, `"technical_values"`) || !strings.Contains(req.ContextPayload, `"HV21"`) {
		t.Fatal("expected technical_values to survive context truncation")
	}
}

func TestPrepareAIContextLoadsMissingData(t *testing.T) {
	provider := &aiPrepProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.config.Watchlist = []string{"AAPL", "SPY"}

	cmd := model.prepareAIContextCmd("ping")
	msg := cmd().(aiContextPreparedMsg)

	if provider.quoteCalls == 0 || provider.fundamentalsCalls == 0 || provider.newsCalls == 0 || len(provider.statementCalls) != len(aiStatementRequests) {
		t.Fatal("expected missing AI context data to be loaded")
	}
	if len(provider.historyCalls) < 2 {
		t.Fatal("expected both chart and technical history loads when missing")
	}
	if msg.quote == nil || msg.fundamentals == nil || msg.technical == nil || msg.history == nil || len(msg.statementBundle) != len(aiStatementRequests) {
		t.Fatal("expected prepared AI context message to contain loaded data")
	}
}

func TestPrepareAIContextSkipsLoadedData(t *testing.T) {
	provider := &aiPrepProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})
	model.quote = domain.QuoteSnapshot{Symbol: model.activeSymbol(), Price: 100}
	model.series = domain.PriceSeries{Symbol: model.activeSymbol(), Candles: make([]domain.Candle, 10)}
	model.technicalCache[strings.ToUpper(model.activeSymbol())] = domain.PriceSeries{Symbol: model.activeSymbol(), Candles: make([]domain.Candle, 252)}
	model.news = []domain.NewsItem{{Title: "Loaded"}}
	model.fundamentals = domain.FundamentalsSnapshot{Symbol: model.activeSymbol(), Description: "Loaded", MarketCap: 1}
	populateAllStatementCache(&model, model.activeSymbol())

	cmd := model.prepareAIContextCmd("ping")
	_ = cmd().(aiContextPreparedMsg)

	if provider.quoteCalls != 0 || provider.fundamentalsCalls != 0 || provider.newsCalls != 0 || len(provider.historyCalls) != 0 || len(provider.statementCalls) != 0 {
		t.Fatal("expected AI context preparation to skip already loaded data")
	}
}

func TestPrepareQuoteInsightLoadsMissingData(t *testing.T) {
	provider := &aiPrepProvider{}
	model := NewModel(context.Background(), Dependencies{
		Config:   storage.DefaultConfig(),
		Registry: providers.NewRegistry(provider),
	})

	cmd := model.prepareQuoteInsightCmd("AAPL")
	msg := cmd().(aiQuoteInsightPreparedMsg)

	if provider.quoteCalls == 0 || provider.fundamentalsCalls == 0 || provider.newsCalls == 0 || len(provider.statementCalls) != len(aiStatementRequests) {
		t.Fatal("expected missing quote insight context data to be loaded")
	}
	if len(provider.historyCalls) < 2 {
		t.Fatal("expected both chart and technical history loads for quote insight")
	}
	if msg.quote == nil || msg.fundamentals == nil || msg.technical == nil || msg.history == nil || len(msg.statementBundle) != len(aiStatementRequests) {
		t.Fatal("expected prepared quote insight message to contain loaded data")
	}
}
