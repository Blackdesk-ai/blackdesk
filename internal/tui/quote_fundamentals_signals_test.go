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
