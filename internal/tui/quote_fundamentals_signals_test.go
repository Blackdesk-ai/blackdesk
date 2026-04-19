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
	rows := quoteFundamentalsProfitabilityRows(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		RevenueGrowth:  0.08,
		EarningsGrowth: -0.03,
		TrailingPE:     30,
		ForwardPE:      24,
		PEGRatio:       2,
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
	if rows[9].name != "N5Y Growth" || !rows[9].styled || rows[9].move <= 0 {
		t.Fatalf("expected n5y growth row to be positively styled, got %+v", rows[9])
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

func TestForwardEarningsYieldValueUsesForwardPE(t *testing.T) {
	got, ok := forwardEarningsYieldValue(domain.FundamentalsSnapshot{ForwardPE: 25})
	if !ok {
		t.Fatal("expected forward earnings yield from forward pe")
	}
	if got != 0.04 {
		t.Fatalf("expected 4%% forward earnings yield, got %.6f", got)
	}
}

func TestFiveYearGrowthEstimateUsesTrailingPEOverPEG(t *testing.T) {
	got, ok := fiveYearGrowthEstimate(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		TrailingPE: 30,
		PEGRatio:   2,
	})
	if !ok {
		t.Fatal("expected 5y growth from trailing pe and peg")
	}
	if got != 0.15 {
		t.Fatalf("expected 15%% 5y growth, got %.6f", got)
	}
}

func TestFiveYearGrowthEstimateFallsBackToQuotePEG(t *testing.T) {
	got, ok := fiveYearGrowthEstimate(domain.QuoteSnapshot{TrailingPEGRatio: 2.5}, domain.FundamentalsSnapshot{
		TrailingPE: 25,
	})
	if !ok {
		t.Fatal("expected 5y growth fallback from quote peg")
	}
	if got != 0.10 {
		t.Fatalf("expected 10%% 5y growth, got %.6f", got)
	}
}

func TestQuoteFundamentalsValuationRowsIncludeForwardEarningsYieldAfterEarningsYield(t *testing.T) {
	rows := quoteFundamentalsValuationRows(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		TrailingPE: 20,
		ForwardPE:  16,
	})

	if rows[4].name != "Earn. Yield" {
		t.Fatalf("expected earnings yield row at index 4, got %+v", rows[4])
	}
	if rows[5].name != "FwdEarn. Yield" {
		t.Fatalf("expected forward earnings yield row after earnings yield, got %+v", rows[5])
	}
	if rows[5].price != "6.25%" {
		t.Fatalf("expected forward earnings yield to render from forward pe, got %+v", rows[5])
	}
}

func TestQuoteFundamentalsProfitabilityRowsIncludeFiveYearGrowthAfterForwardGrowth(t *testing.T) {
	rows := quoteFundamentalsProfitabilityRows(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		TrailingPE: 30,
		ForwardPE:  24,
		PEGRatio:   2,
	})

	if rows[8].name != "Fwd Growth" {
		t.Fatalf("expected forward growth row at index 8, got %+v", rows[8])
	}
	if rows[9].name != "N5Y Growth" {
		t.Fatalf("expected n5y growth row after forward growth, got %+v", rows[9])
	}
	if rows[9].price != "15.00%" {
		t.Fatalf("expected n5y growth to render from pe over peg, got %+v", rows[9])
	}
}

func TestFCFYieldValueUsesFreeCashFlowOverMarketCap(t *testing.T) {
	got, ok := fcfYieldValue(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		MarketCap:    100_000_000_000,
		FreeCashflow: 5_000_000_000,
	})
	if !ok {
		t.Fatal("expected fcf yield from market cap and free cash flow")
	}
	if got != 0.05 {
		t.Fatalf("expected 5%% fcf yield, got %.6f", got)
	}
}

func TestFCFYieldValueFallsBackToQuoteMarketCap(t *testing.T) {
	got, ok := fcfYieldValue(domain.QuoteSnapshot{MarketCap: 80_000_000_000}, domain.FundamentalsSnapshot{
		FreeCashflow: 4_000_000_000,
	})
	if !ok {
		t.Fatal("expected fcf yield fallback from quote market cap")
	}
	if got != 0.05 {
		t.Fatalf("expected 5%% fcf yield, got %.6f", got)
	}
}

func TestValuationScoreValueMultipliesEarningsYieldAndROIC(t *testing.T) {
	got, ok := valuationScoreValue(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		TrailingPE:              25,
		ReturnOnInvestedCapital: 0.18,
	})
	if !ok {
		t.Fatal("expected valuation score when earnings yield and roic are available")
	}
	if got != 0.0072 {
		t.Fatalf("expected multiplied score, got %.6f", got)
	}
}

func TestValuationScoreValueTurnsNegativeWhenBothInputsAreNegative(t *testing.T) {
	got, ok := valuationScoreValue(domain.QuoteSnapshot{Price: 100}, domain.FundamentalsSnapshot{
		TrailingEPS:             -11.7,
		ReturnOnInvestedCapital: -0.4936,
	})
	if !ok {
		t.Fatal("expected valuation score when both inputs are available")
	}
	if got >= 0 {
		t.Fatalf("expected negative score for double-negative inputs, got %.6f", got)
	}
}

func TestValuationScoreValueTurnsNegativeWhenOnlyOneInputIsNegative(t *testing.T) {
	got, ok := valuationScoreValue(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		TrailingPE:              25,
		ReturnOnInvestedCapital: -0.18,
	})
	if !ok {
		t.Fatal("expected valuation score when earnings yield and roic are available")
	}
	if got >= 0 {
		t.Fatalf("expected negative score when one input is negative, got %.6f", got)
	}
}

func TestValuationScoreValueRequiresROIC(t *testing.T) {
	_, ok := valuationScoreValue(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{TrailingPE: 25})
	if ok {
		t.Fatal("expected missing roic to disable valuation score")
	}
}

func TestRuleOf40ValueAddsRevenueGrowthAndProfitMargin(t *testing.T) {
	got, ok := ruleOf40Value(domain.FundamentalsSnapshot{
		RevenueGrowth: 0.061,
		ProfitMargins: 0.262,
	})
	if !ok {
		t.Fatal("expected rule of 40 score when growth or margin is available")
	}
	if got != 0.323 {
		t.Fatalf("expected 32.3%% r40 score, got %.6f", got)
	}
}

func TestRuleOf40ValueRequiresAtLeastOneInput(t *testing.T) {
	_, ok := ruleOf40Value(domain.FundamentalsSnapshot{})
	if ok {
		t.Fatal("expected missing growth and margin to disable r40 score")
	}
}

func TestQuoteFundamentalsValuationRowsDoNotIncludeQARPScore(t *testing.T) {
	rows := quoteFundamentalsValuationRows(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		MarketCap:               100,
		EnterpriseValue:         120,
		TrailingPE:              25,
		ReturnOnInvestedCapital: 0.18,
	})

	for _, row := range rows {
		if row.name == "QARP Score" || row.name == "R40" {
			t.Fatalf("expected supplemental scores to be rendered outside valuation rows, got %+v", row)
		}
	}
}
