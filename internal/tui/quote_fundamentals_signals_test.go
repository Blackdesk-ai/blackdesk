package tui

import (
	"testing"

	"blackdesk/internal/domain"
)

func TestImpliedEPSGrowthEstimateTextUsesTrailingAndForwardPE(t *testing.T) {
	got := impliedEPSGrowthEstimateText(domain.FundamentalsSnapshot{
		TrailingPE: 31.2,
		ForwardPE:  28.4,
	})
	if got != "10%" {
		t.Fatalf("expected implied growth estimate, got %q", got)
	}
}

func TestImpliedEPSGrowthEstimateTextRequiresTrailingAndForwardPE(t *testing.T) {
	got := impliedEPSGrowthEstimateText(domain.FundamentalsSnapshot{TrailingPE: 30})
	if got != "-" {
		t.Fatalf("expected missing forward pe to disable growth estimate, got %q", got)
	}
}

func TestQuoteFundamentalsProfitabilityRowsStylesGrowthValues(t *testing.T) {
	rows := quoteFundamentalsProfitabilityRows(domain.FundamentalsSnapshot{
		RevenueGrowth:  0.08,
		EarningsGrowth: -0.03,
		TrailingPE:     30,
		ForwardPE:      24,
	})

	if rows[6].name != "Rev growth" || !rows[6].styled || rows[6].move <= 0 {
		t.Fatalf("expected revenue growth row to be positively styled, got %+v", rows[6])
	}
	if rows[7].name != "EPS growth" || !rows[7].styled || rows[7].move >= 0 {
		t.Fatalf("expected eps growth row to be negatively styled, got %+v", rows[7])
	}
	if rows[8].name != "Fwd Growth" || !rows[8].styled || rows[8].move <= 0 {
		t.Fatalf("expected forward growth row to be positively styled, got %+v", rows[8])
	}
}

func TestEarningsYieldValueUsesTrailingPE(t *testing.T) {
	got, ok := earningsYieldValue(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{TrailingPE: 31.2})
	if !ok {
		t.Fatal("expected earnings yield from trailing pe")
	}
	if got < 0.0320 || got > 0.0321 {
		t.Fatalf("expected earnings yield near 3.21%%, got %.6f", got)
	}
}

func TestEarningsYieldValueFallsBackToEPSOverPrice(t *testing.T) {
	got, ok := earningsYieldValue(domain.QuoteSnapshot{Price: 200}, domain.FundamentalsSnapshot{TrailingEPS: 8})
	if !ok {
		t.Fatal("expected earnings yield fallback from eps and price")
	}
	if got != 0.04 {
		t.Fatalf("expected 4%% earnings yield, got %.6f", got)
	}
}
