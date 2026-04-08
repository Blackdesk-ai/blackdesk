package yahoo

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

const (
	defaultScreenerCount = 25
	maxScreenerCount     = 50
)

var screenerCatalog = []struct {
	def      domain.ScreenerDefinition
	criteria []domain.ScreenerCriterion
}{
	{
		def: domain.ScreenerDefinition{
			ID:          "most_actives",
			Name:        "Most Active",
			Description: "Most traded US equities with meaningful scale and turnover.",
			Category:    "Market Movers",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "region", Operator: "eq", Value: "US", Statement: "Region = US"},
			{Field: "market_cap", Operator: "gte", Value: "$2B", Statement: "Market cap >= $2B"},
			{Field: "day_volume", Operator: "gt", Value: "5M", Statement: "Day volume > 5M"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "day_gainers",
			Name:        "Day Gainers",
			Description: "US names posting strong upside on the session.",
			Category:    "Market Movers",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "percent_change", Operator: "gt", Value: "3%", Statement: "Session change > 3%"},
			{Field: "region", Operator: "eq", Value: "US", Statement: "Region = US"},
			{Field: "market_cap", Operator: "gte", Value: "$2B", Statement: "Market cap >= $2B"},
			{Field: "price", Operator: "gte", Value: "$5", Statement: "Price >= $5"},
			{Field: "day_volume", Operator: "gt", Value: "15K", Statement: "Day volume > 15K"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "day_losers",
			Name:        "Day Losers",
			Description: "US names under the most pressure during the session.",
			Category:    "Market Movers",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "percent_change", Operator: "lt", Value: "-2.5%", Statement: "Session change < -2.5%"},
			{Field: "region", Operator: "eq", Value: "US", Statement: "Region = US"},
			{Field: "market_cap", Operator: "gte", Value: "$2B", Statement: "Market cap >= $2B"},
			{Field: "price", Operator: "gte", Value: "$5", Statement: "Price >= $5"},
			{Field: "day_volume", Operator: "gt", Value: "20K", Statement: "Day volume > 20K"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "most_shorted_stocks",
			Name:        "Most Shorted",
			Description: "Liquid US equities ranked by elevated short-interest pressure.",
			Category:    "Market Movers",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "region", Operator: "eq", Value: "US", Statement: "Region = US"},
			{Field: "price", Operator: "gt", Value: "$1", Statement: "Price > $1"},
			{Field: "avg_3m_volume", Operator: "gt", Value: "200K", Statement: "Avg 3M volume > 200K"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "aggressive_small_caps",
			Name:        "Aggressive Small Caps",
			Description: "Smaller-cap exchange names sorted for active opportunity hunting.",
			Category:    "Market Movers",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "exchange", Operator: "in", Value: "NMS, NYQ", Statement: "Exchange in NMS / NYQ"},
			{Field: "eps_growth_ttm", Operator: "lt", Value: "15", Statement: "EPS growth TTM < 15"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "small_cap_gainers",
			Name:        "Small Cap Gainers",
			Description: "Smaller-cap names with active tape and upside follow-through.",
			Category:    "Market Movers",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "market_cap", Operator: "lt", Value: "$2B", Statement: "Market cap < $2B"},
			{Field: "exchange", Operator: "in", Value: "NMS, NYQ", Statement: "Exchange in NMS / NYQ"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "growth_technology_stocks",
			Name:        "Growth Technology",
			Description: "Technology names with strong revenue and EPS growth filters.",
			Category:    "Value & Growth",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "quarterly_revenue_growth", Operator: "gte", Value: "25%", Statement: "Quarterly revenue growth >= 25%"},
			{Field: "eps_growth_ttm", Operator: "gte", Value: "25%", Statement: "EPS growth TTM >= 25%"},
			{Field: "sector", Operator: "eq", Value: "Technology", Statement: "Sector = Technology"},
			{Field: "exchange", Operator: "in", Value: "NMS, NYQ", Statement: "Exchange in NMS / NYQ"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "undervalued_growth_stocks",
			Name:        "Undervalued Growth",
			Description: "Growth names screened for lower valuation relative to expansion.",
			Category:    "Value & Growth",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "pe_ttm", Operator: "between", Value: "0-20", Statement: "P/E TTM between 0 and 20"},
			{Field: "peg_5y", Operator: "lt", Value: "1", Statement: "PEG 5Y < 1"},
			{Field: "eps_growth_ttm", Operator: "gte", Value: "25%", Statement: "EPS growth TTM >= 25%"},
			{Field: "exchange", Operator: "in", Value: "NMS, NYQ", Statement: "Exchange in NMS / NYQ"},
		},
	},
	{
		def: domain.ScreenerDefinition{
			ID:          "undervalued_large_caps",
			Name:        "Undervalued Large Caps",
			Description: "Larger-cap names screened for lower valuation and PEG discipline.",
			Category:    "Value & Growth",
			Kind:        "equity",
		},
		criteria: []domain.ScreenerCriterion{
			{Field: "pe_ttm", Operator: "between", Value: "0-20", Statement: "P/E TTM between 0 and 20"},
			{Field: "peg_5y", Operator: "lt", Value: "1", Statement: "PEG 5Y < 1"},
			{Field: "market_cap", Operator: "between", Value: "$10B-$100B", Statement: "Market cap between $10B and $100B"},
			{Field: "exchange", Operator: "in", Value: "NMS, NYQ", Statement: "Exchange in NMS / NYQ"},
		},
	},
}

type screenerResponse struct {
	Finance struct {
		Result []struct {
			Title         string `json:"title"`
			Description   string `json:"description"`
			CanonicalName string `json:"canonicalName"`
			Start         int    `json:"start"`
			Count         int    `json:"count"`
			Total         int    `json:"total"`
			CriteriaMeta  struct {
				SortField string `json:"sortField"`
				SortType  string `json:"sortType"`
				QuoteType string `json:"quoteType"`
			} `json:"criteriaMeta"`
			Quotes []screenerQuote `json:"quotes"`
		} `json:"result"`
	} `json:"finance"`
}

type screenerQuote struct {
	Symbol                     string  `json:"symbol"`
	ShortName                  string  `json:"shortName"`
	LongName                   string  `json:"longName"`
	FullExchangeName           string  `json:"fullExchangeName"`
	Exchange                   string  `json:"exchange"`
	QuoteType                  string  `json:"quoteType"`
	TypeDisp                   string  `json:"typeDisp"`
	Currency                   string  `json:"currency"`
	MarketState                string  `json:"marketState"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketVolume        int64   `json:"regularMarketVolume"`
	AverageDailyVolume3Month   int64   `json:"averageDailyVolume3Month"`
	MarketCap                  int64   `json:"marketCap"`
	RegularMarketTime          int64   `json:"regularMarketTime"`
	TrailingPE                 float64 `json:"trailingPE"`
	ForwardPE                  float64 `json:"forwardPE"`
	PriceToBook                float64 `json:"priceToBook"`
	FiftyTwoWeekHigh           float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            float64 `json:"fiftyTwoWeekLow"`
	FiftyTwoWeekChangePercent  float64 `json:"fiftyTwoWeekChangePercent"`
	DividendYield              float64 `json:"dividendYield"`
	AverageAnalystRating       string  `json:"averageAnalystRating"`
	NetAssets                  float64 `json:"netAssets"`
	NetExpenseRatio            float64 `json:"netExpenseRatio"`
	YTDReturn                  float64 `json:"ytdReturn"`
	TrailingThreeMonthReturns  float64 `json:"trailingThreeMonthReturns"`
	AnnualReturnNavY1          float64 `json:"annualReturnNavY1"`
	AnnualReturnNavY3          float64 `json:"annualReturnNavY3"`
	AnnualReturnNavY5          float64 `json:"annualReturnNavY5"`
	YieldTTM                   float64 `json:"yieldTTM"`
	PerformanceRatingOverall   float64 `json:"performanceRatingOverall"`
	RiskRatingOverall          float64 `json:"riskRatingOverall"`
}

func (p *Provider) Screeners() []domain.ScreenerDefinition {
	out := make([]domain.ScreenerDefinition, 0, len(screenerCatalog))
	for _, item := range screenerCatalog {
		out = append(out, item.def)
	}
	return out
}

func (p *Provider) GetScreener(ctx context.Context, id string, count int) (domain.ScreenerResult, error) {
	spec, ok := lookupScreener(id)
	if !ok {
		return domain.ScreenerResult{}, fmt.Errorf("screener %q not supported", strings.TrimSpace(id))
	}

	count = clampScreenerCount(count)

	var resp screenerResponse
	params := url.Values{}
	params.Set("formatted", "false")
	params.Set("lang", "en-US")
	params.Set("region", "US")
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("scrIds", spec.def.ID)
	params.Set("count", strconv.Itoa(count))

	err := p.fetchJSON(ctx, requestSpec{
		URL:      "https://query2.finance.yahoo.com/v1/finance/screener/predefined/saved",
		Params:   params,
		CacheKey: "screener:" + spec.def.ID + ":" + strconv.Itoa(count),
		TTL:      2 * time.Minute,
		Auth:     authOptional,
	}, &resp)
	if err != nil {
		return domain.ScreenerResult{}, err
	}
	return normalizeScreener(resp, spec)
}

func lookupScreener(id string) (struct {
	def      domain.ScreenerDefinition
	criteria []domain.ScreenerCriterion
}, bool) {
	needle := strings.ToLower(strings.TrimSpace(id))
	for _, item := range screenerCatalog {
		if item.def.ID == needle {
			return item, true
		}
	}
	return struct {
		def      domain.ScreenerDefinition
		criteria []domain.ScreenerCriterion
	}{}, false
}

func clampScreenerCount(count int) int {
	if count <= 0 {
		return defaultScreenerCount
	}
	if count > maxScreenerCount {
		return maxScreenerCount
	}
	return count
}

func normalizeScreener(resp screenerResponse, spec struct {
	def      domain.ScreenerDefinition
	criteria []domain.ScreenerCriterion
}) (domain.ScreenerResult, error) {
	if len(resp.Finance.Result) == 0 {
		return domain.ScreenerResult{}, errors.New("screener not found")
	}
	raw := resp.Finance.Result[0]
	def := spec.def
	if strings.TrimSpace(raw.Title) != "" {
		def.Name = strings.TrimSpace(raw.Title)
	}
	if strings.TrimSpace(raw.Description) != "" {
		def.Description = strings.TrimSpace(raw.Description)
	}

	items := make([]domain.ScreenerItem, 0, len(raw.Quotes))
	for _, quote := range raw.Quotes {
		item := domain.ScreenerItem{
			Symbol:        strings.ToUpper(strings.TrimSpace(quote.Symbol)),
			Name:          firstNonEmpty(strings.TrimSpace(quote.ShortName), strings.TrimSpace(quote.LongName), strings.ToUpper(strings.TrimSpace(quote.Symbol))),
			Exchange:      firstNonEmpty(strings.TrimSpace(quote.FullExchangeName), strings.TrimSpace(quote.Exchange)),
			Type:          firstNonEmpty(strings.TrimSpace(quote.TypeDisp), strings.TrimSpace(quote.QuoteType)),
			Currency:      strings.TrimSpace(quote.Currency),
			Price:         quote.RegularMarketPrice,
			Change:        quote.RegularMarketChange,
			ChangePercent: quote.RegularMarketChangePercent,
			Volume:        quote.RegularMarketVolume,
			AverageVolume: quote.AverageDailyVolume3Month,
			MarketCap:     quote.MarketCap,
			MarketState:   normalizeMarketState(quote.MarketState),
			UpdatedAt:     time.Unix(quote.RegularMarketTime, 0),
			Metrics:       buildScreenerMetrics(spec.def.ID, spec.def.Kind, quote),
		}
		items = append(items, item)
	}

	return domain.ScreenerResult{
		Definition: def,
		SortField:  strings.TrimSpace(raw.CriteriaMeta.SortField),
		SortOrder:  strings.TrimSpace(raw.CriteriaMeta.SortType),
		Total:      raw.Total,
		Items:      items,
		Criteria:   append([]domain.ScreenerCriterion(nil), spec.criteria...),
		Freshness:  domain.FreshnessLive,
		Provider:   "yahoo",
		UpdatedAt:  time.Now(),
	}, nil
}

func buildScreenerMetrics(id, kind string, quote screenerQuote) []domain.ScreenerMetric {
	metrics := make([]domain.ScreenerMetric, 0, 6)
	add := func(key, label, value, signal string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		metrics = append(metrics, domain.ScreenerMetric{
			Key:    key,
			Label:  label,
			Value:  value,
			Signal: signal,
		})
	}

	if kind == "fund" {
		add("relative_volume", "RV", formatRelativeVolume(quote.RegularMarketVolume, quote.AverageDailyVolume3Month), "")
		add("return_1y", "1Y Return", formatPercentValue(firstNonZero(quote.AnnualReturnNavY1, quote.YTDReturn)), metricSignalFromPercent(firstNonZero(quote.AnnualReturnNavY1, quote.YTDReturn)))
		add("return_3y", "3Y Return", formatPercentValue(quote.AnnualReturnNavY3), metricSignalFromPercent(quote.AnnualReturnNavY3))
		add("return_5y", "5Y Return", formatPercentValue(quote.AnnualReturnNavY5), metricSignalFromPercent(quote.AnnualReturnNavY5))
		add("net_assets", "Net Assets", formatCompactFloat(quote.NetAssets), "")
		add("expense_ratio", "Expense", formatRatio(quote.NetExpenseRatio), "")
		add("yield_ttm", "Yield TTM", formatPercentValue(quote.YieldTTM), metricSignalFromPercent(quote.YieldTTM))
		add("perf_rating", "Perf Rating", formatRating5(quote.PerformanceRatingOverall), "")
		add("risk_rating", "Risk Rating", formatRating5(quote.RiskRatingOverall), "")
		add("return_3m", "3M Return", formatPercentValue(quote.TrailingThreeMonthReturns), metricSignalFromPercent(quote.TrailingThreeMonthReturns))
		return metrics
	}

	switch id {
	case "most_actives", "most_shorted_stocks":
		add("avg_3m_volume", "Avg 3M Vol", formatCompactInt(quote.AverageDailyVolume3Month), "")
	case "day_gainers", "day_losers", "small_cap_gainers":
		add("fifty_two_week_change", "52W Change", formatPercentValue(quote.FiftyTwoWeekChangePercent), metricSignalFromPercent(quote.FiftyTwoWeekChangePercent))
	case "growth_technology_stocks", "undervalued_growth_stocks", "undervalued_large_caps":
		add("forward_pe", "Fwd P/E", formatDecimal(quote.ForwardPE), metricSignalFromInverseRatio(quote.ForwardPE))
	}
	add("relative_volume", "RV", formatRelativeVolume(quote.RegularMarketVolume, quote.AverageDailyVolume3Month), "")
	add("trailing_pe", "P/E TTM", formatDecimal(quote.TrailingPE), metricSignalFromInverseRatio(quote.TrailingPE))
	add("price_to_book", "P/B", formatDecimal(quote.PriceToBook), metricSignalFromInverseRatio(quote.PriceToBook))
	add("dividend_yield", "Div Yield", formatPercentValue(quote.DividendYield), metricSignalFromPercent(quote.DividendYield))
	add("analyst_rating", "Analyst", strings.TrimSpace(quote.AverageAnalystRating), "")
	add("avg_3m_volume", "Avg 3M Vol", formatCompactInt(quote.AverageDailyVolume3Month), "")
	add("fifty_two_week_change", "52W Change", formatPercentValue(quote.FiftyTwoWeekChangePercent), metricSignalFromPercent(quote.FiftyTwoWeekChangePercent))
	return dedupeScreenerMetrics(metrics)
}

func dedupeScreenerMetrics(metrics []domain.ScreenerMetric) []domain.ScreenerMetric {
	out := make([]domain.ScreenerMetric, 0, len(metrics))
	seen := make(map[string]struct{}, len(metrics))
	for _, metric := range metrics {
		if metric.Key == "" {
			continue
		}
		if _, ok := seen[metric.Key]; ok {
			continue
		}
		seen[metric.Key] = struct{}{}
		out = append(out, metric)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func formatCompactInt(v int64) string {
	if v == 0 {
		return ""
	}
	return formatCompactFloat(float64(v))
}

func formatCompactFloat(v float64) string {
	if v == 0 {
		return ""
	}
	abs := v
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs >= 1_000_000_000_000:
		return fmt.Sprintf("%.1fT", v/1_000_000_000_000)
	case abs >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", v/1_000_000_000)
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fK", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

func formatRelativeVolume(volume, averageVolume int64) string {
	if volume <= 0 || averageVolume <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2fx", float64(volume)/float64(averageVolume))
}

func formatPercentValue(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%+.1f%%", v)
}

func formatRatio(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f%%", v)
}

func formatDecimal(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", v)
}

func formatRating5(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.0f/5", v)
}

func metricSignalFromPercent(v float64) string {
	switch {
	case v > 0:
		return "up"
	case v < 0:
		return "down"
	default:
		return ""
	}
}

func metricSignalFromInverseRatio(v float64) string {
	switch {
	case v > 0 && v <= 15:
		return "low"
	case v >= 30:
		return "high"
	default:
		return ""
	}
}
