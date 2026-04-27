package tui

import (
	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
	"fmt"
	"math"
)

func formatCompactIntDash(v int64) string {
	if v == 0 {
		return "-"
	}
	return ui.FormatCompactInt(v)
}

func cashFlowSignal(v int64) string {
	switch {
	case v > 0:
		return "positive"
	case v < 0:
		return "burn"
	default:
		return "-"
	}
}

func netCashSignal(v int64) string {
	switch {
	case v > 0:
		return "net cash"
	case v < 0:
		return "net debt"
	default:
		return "-"
	}
}

func analystUpsideValue(f domain.FundamentalsSnapshot, price float64) float64 {
	if price == 0 || f.TargetMeanPrice == 0 {
		return 0
	}
	return (f.TargetMeanPrice - price) / price
}

func pegRatioValue(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) float64 {
	if f.PEGRatio != 0 {
		return f.PEGRatio
	}
	return q.TrailingPEGRatio
}

func earningsYieldValue(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) (float64, bool) {
	if f.TrailingPE != 0 {
		return 1 / f.TrailingPE, true
	}
	if q.Price <= 0 {
		return 0, false
	}
	eps := f.TrailingEPS
	if eps == 0 {
		eps = f.EPS
	}
	if eps == 0 {
		return 0, false
	}
	return eps / q.Price, true
}

func forwardEarningsYieldValue(f domain.FundamentalsSnapshot) (float64, bool) {
	if f.ForwardPE == 0 {
		return 0, false
	}
	return 1 / f.ForwardPE, true
}

func fcfYieldValue(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) (float64, bool) {
	marketCap := f.MarketCap
	if marketCap == 0 {
		marketCap = q.MarketCap
	}
	if marketCap <= 0 || f.FreeCashflow == 0 {
		return 0, false
	}
	return float64(f.FreeCashflow) / float64(marketCap), true
}

func valuationScoreValue(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) (float64, bool) {
	earningsYield, earningsYieldOK := earningsYieldValue(q, f)
	if !earningsYieldOK || f.ReturnOnInvestedCapital == 0 {
		return 0, false
	}
	score := earningsYield * f.ReturnOnInvestedCapital
	if earningsYield < 0 || f.ReturnOnInvestedCapital < 0 {
		return -math.Abs(score), true
	}
	return score, true
}

func impliedReturnValue(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) (float64, bool) {
	earningsYield, earningsYieldOK := earningsYieldValue(q, f)
	fiveYearGrowth, fiveYearGrowthOK := fiveYearGrowthEstimate(q, f)
	if !earningsYieldOK || !fiveYearGrowthOK {
		return 0, false
	}
	return earningsYield + fiveYearGrowth, true
}

func impliedSharpeValue(impliedReturn, hv252 float64) (float64, bool) {
	if hv252 <= 0 {
		return 0, false
	}
	return impliedReturn / hv252, true
}

func impliedEPSGrowthEstimate(f domain.FundamentalsSnapshot) (float64, bool) {
	if f.TrailingPE <= 0 || f.ForwardPE <= 0 {
		return 0, false
	}
	return f.TrailingPE/f.ForwardPE - 1, true
}

func fiveYearGrowthEstimate(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) (float64, bool) {
	peg := pegRatioValue(q, f)
	if f.TrailingPE == 0 || peg == 0 {
		return 0, false
	}
	return (f.TrailingPE / peg) / 100, true
}

func impliedEPSGrowthEstimateText(f domain.FundamentalsSnapshot) string {
	growth, ok := impliedEPSGrowthEstimate(f)
	if !ok {
		return "-"
	}
	return formatImpliedGrowthPercent(growth * 100)
}

func formatImpliedGrowthPercent(v float64) string {
	return fmt.Sprintf("%.0f%%", v)
}
