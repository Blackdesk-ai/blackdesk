package yahoo

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"blackdesk/internal/domain"
)

type analystRecommendationsResponse struct {
	QuoteSummary struct {
		Result []struct {
			Price struct {
				ShortName string `json:"shortName"`
				LongName  string `json:"longName"`
			} `json:"price"`
			FinancialData struct {
				RecommendationKey       string      `json:"recommendationKey"`
				RecommendationMean      numberField `json:"recommendationMean"`
				NumberOfAnalystOpinions numberField `json:"numberOfAnalystOpinions"`
				TargetLowPrice          numberField `json:"targetLowPrice"`
				TargetMeanPrice         numberField `json:"targetMeanPrice"`
				TargetHighPrice         numberField `json:"targetHighPrice"`
			} `json:"financialData"`
			RecommendationTrend struct {
				Trend []struct {
					Period     string `json:"period"`
					StrongBuy  int    `json:"strongBuy"`
					Buy        int    `json:"buy"`
					Hold       int    `json:"hold"`
					Sell       int    `json:"sell"`
					StrongSell int    `json:"strongSell"`
				} `json:"trend"`
			} `json:"recommendationTrend"`
			UpgradeDowngradeHistory struct {
				History []struct {
					EpochGradeDate int64  `json:"epochGradeDate"`
					Firm           string `json:"firm"`
					ToGrade        string `json:"toGrade"`
					FromGrade      string `json:"fromGrade"`
					Action         string `json:"action"`
				} `json:"history"`
			} `json:"upgradeDowngradeHistory"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

func (p *Provider) GetAnalystRecommendations(ctx context.Context, symbol string) (domain.AnalystRecommendationsSnapshot, error) {
	var resp analystRecommendationsResponse
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	params := url.Values{}
	params.Set("modules", "price,financialData,recommendationTrend,upgradeDowngradeHistory")
	params.Set("corsDomain", "finance.yahoo.com")
	params.Set("formatted", "false")
	params.Set("symbol", normalizedSymbol)
	err := p.fetchJSON(ctx, requestSpec{
		URL:      p.quoteSummaryBase + url.PathEscape(normalizedSymbol),
		Params:   params,
		CacheKey: "analyst:" + normalizedSymbol,
		TTL:      30 * time.Minute,
		Auth:     authRequired,
	}, &resp)
	if err != nil {
		return domain.AnalystRecommendationsSnapshot{}, err
	}
	return normalizeAnalystRecommendations(normalizedSymbol, resp)
}

func normalizeAnalystRecommendations(symbol string, resp analystRecommendationsResponse) (domain.AnalystRecommendationsSnapshot, error) {
	if len(resp.QuoteSummary.Result) == 0 {
		return domain.AnalystRecommendationsSnapshot{}, fmt.Errorf("yahoo analyst recommendations unavailable")
	}
	result := resp.QuoteSummary.Result[0]
	snapshot := domain.AnalystRecommendationsSnapshot{
		Symbol:             symbol,
		CompanyName:        firstNonEmptyString(result.Price.ShortName, result.Price.LongName),
		RecommendationKey:  strings.TrimSpace(result.FinancialData.RecommendationKey),
		RecommendationMean: result.FinancialData.RecommendationMean.Raw,
		AnalystOpinions:    int(math.Round(result.FinancialData.NumberOfAnalystOpinions.Raw)),
		TargetLowPrice:     result.FinancialData.TargetLowPrice.Raw,
		TargetMeanPrice:    result.FinancialData.TargetMeanPrice.Raw,
		TargetHighPrice:    result.FinancialData.TargetHighPrice.Raw,
		Freshness:          domain.FreshnessLive,
		Provider:           "yahoo",
		UpdatedAt:          time.Now(),
	}

	for _, item := range result.UpgradeDowngradeHistory.History {
		snapshot.Items = append(snapshot.Items, domain.AnalystRecommendationItem{
			Firm:      strings.TrimSpace(item.Firm),
			Action:    strings.TrimSpace(item.Action),
			ToGrade:   normalizeRecommendationGrade(item.ToGrade),
			FromGrade: normalizeRecommendationGrade(item.FromGrade),
			Date:      normalizeEpochDate(item.EpochGradeDate),
		})
	}
	sort.SliceStable(snapshot.Items, func(i, j int) bool {
		if snapshot.Items[i].Date.Equal(snapshot.Items[j].Date) {
			return snapshot.Items[i].Firm < snapshot.Items[j].Firm
		}
		return snapshot.Items[i].Date.After(snapshot.Items[j].Date)
	})

	for _, trend := range result.RecommendationTrend.Trend {
		snapshot.Trends = append(snapshot.Trends, domain.AnalystRecommendationTrend{
			Period:     strings.TrimSpace(trend.Period),
			StrongBuy:  trend.StrongBuy,
			Buy:        trend.Buy,
			Hold:       trend.Hold,
			Sell:       trend.Sell,
			StrongSell: trend.StrongSell,
		})
	}
	sort.SliceStable(snapshot.Trends, func(i, j int) bool {
		return analystTrendOrder(snapshot.Trends[i].Period) < analystTrendOrder(snapshot.Trends[j].Period)
	})

	if snapshot.CompanyName == "" {
		snapshot.CompanyName = snapshot.Symbol
	}
	return snapshot, nil
}

func normalizeRecommendationGrade(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return strings.Join(strings.Fields(raw), " ")
}

func normalizeEpochDate(raw int64) time.Time {
	if raw <= 0 {
		return time.Time{}
	}
	return time.Unix(raw, 0).UTC()
}

func analystTrendOrder(period string) int {
	switch strings.TrimSpace(period) {
	case "0m":
		return 0
	case "-1m":
		return 1
	case "-2m":
		return 2
	case "-3m":
		return 3
	default:
		return 10
	}
}
