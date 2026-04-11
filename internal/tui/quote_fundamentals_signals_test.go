package tui

import (
	"testing"

	"blackdesk/internal/domain"
)

func TestImpliedEPSGrowthBandTextUsesForwardAndTrailingPE(t *testing.T) {
	got := impliedEPSGrowthBandText(domain.QuoteSnapshot{}, domain.FundamentalsSnapshot{
		TrailingPE: 31.2,
		ForwardPE:  28.4,
		PEGRatio:   2.14,
	})
	if got != "13%-15%" {
		t.Fatalf("expected implied growth band, got %q", got)
	}
}

func TestImpliedEPSGrowthBandTextFallsBackToQuoteTrailingPEG(t *testing.T) {
	got := impliedEPSGrowthBandText(domain.QuoteSnapshot{TrailingPEGRatio: 1.5}, domain.FundamentalsSnapshot{
		TrailingPE: 30,
		ForwardPE:  24,
	})
	if got != "16%-20%" {
		t.Fatalf("expected quote trailing peg fallback band, got %q", got)
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
