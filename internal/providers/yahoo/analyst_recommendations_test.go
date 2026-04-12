package yahoo

import (
	"strings"
	"testing"
)

func TestNormalizeAnalystRecommendationsSortsLatestFirst(t *testing.T) {
	resp := analystRecommendationsResponse{}
	resp.QuoteSummary.Result = []struct {
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
	}{
		{
			Price: struct {
				ShortName string `json:"shortName"`
				LongName  string `json:"longName"`
			}{ShortName: "Apple Inc."},
			FinancialData: struct {
				RecommendationKey       string      `json:"recommendationKey"`
				RecommendationMean      numberField `json:"recommendationMean"`
				NumberOfAnalystOpinions numberField `json:"numberOfAnalystOpinions"`
				TargetLowPrice          numberField `json:"targetLowPrice"`
				TargetMeanPrice         numberField `json:"targetMeanPrice"`
				TargetHighPrice         numberField `json:"targetHighPrice"`
			}{
				RecommendationKey:       "buy",
				RecommendationMean:      numberField{Raw: 1.8},
				NumberOfAnalystOpinions: numberField{Raw: 40},
				TargetLowPrice:          numberField{Raw: 190},
				TargetMeanPrice:         numberField{Raw: 220},
				TargetHighPrice:         numberField{Raw: 250},
			},
			RecommendationTrend: struct {
				Trend []struct {
					Period     string `json:"period"`
					StrongBuy  int    `json:"strongBuy"`
					Buy        int    `json:"buy"`
					Hold       int    `json:"hold"`
					Sell       int    `json:"sell"`
					StrongSell int    `json:"strongSell"`
				} `json:"trend"`
			}{
				Trend: []struct {
					Period     string `json:"period"`
					StrongBuy  int    `json:"strongBuy"`
					Buy        int    `json:"buy"`
					Hold       int    `json:"hold"`
					Sell       int    `json:"sell"`
					StrongSell int    `json:"strongSell"`
				}{
					{Period: "-1m", StrongBuy: 10, Buy: 20, Hold: 9, Sell: 1, StrongSell: 0},
					{Period: "0m", StrongBuy: 11, Buy: 20, Hold: 8, Sell: 1, StrongSell: 0},
				},
			},
			UpgradeDowngradeHistory: struct {
				History []struct {
					EpochGradeDate int64  `json:"epochGradeDate"`
					Firm           string `json:"firm"`
					ToGrade        string `json:"toGrade"`
					FromGrade      string `json:"fromGrade"`
					Action         string `json:"action"`
				} `json:"history"`
			}{
				History: []struct {
					EpochGradeDate int64  `json:"epochGradeDate"`
					Firm           string `json:"firm"`
					ToGrade        string `json:"toGrade"`
					FromGrade      string `json:"fromGrade"`
					Action         string `json:"action"`
				}{
					{EpochGradeDate: 1712832000, Firm: "Barclays", ToGrade: "Equal Weight", FromGrade: "Overweight", Action: "down"},
					{EpochGradeDate: 1713004800, Firm: "Morgan Stanley", ToGrade: "Overweight", FromGrade: "Equal-Weight", Action: "up"},
				},
			},
		},
	}

	got, err := normalizeAnalystRecommendations("AAPL", resp)
	if err != nil {
		t.Fatalf("normalizeAnalystRecommendations error: %v", err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 analyst items, got %+v", got.Items)
	}
	if got.Items[0].Firm != "Morgan Stanley" {
		t.Fatalf("expected latest item first, got %+v", got.Items)
	}
	if len(got.Trends) != 2 || got.Trends[0].Period != "0m" {
		t.Fatalf("expected trend periods sorted newest first, got %+v", got.Trends)
	}
	if got.RecommendationKey != "buy" || got.AnalystOpinions != 40 {
		t.Fatalf("unexpected analyst summary %+v", got)
	}
}

func TestNormalizeRecommendationGradeCollapsesWhitespace(t *testing.T) {
	if got := normalizeRecommendationGrade("  Equal   Weight "); got != "Equal Weight" {
		t.Fatalf("expected normalized grade, got %q", got)
	}
}

func TestAnalystTrendOrderPlacesCurrentMonthFirst(t *testing.T) {
	if !(analystTrendOrder("0m") < analystTrendOrder("-1m") && analystTrendOrder("-1m") < analystTrendOrder("-2m")) {
		t.Fatal("expected current month ordering before historical periods")
	}
}

func TestNormalizeAnalystRecommendationsRequiresResult(t *testing.T) {
	_, err := normalizeAnalystRecommendations("AAPL", analystRecommendationsResponse{})
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}
