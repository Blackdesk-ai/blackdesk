package tui

import (
	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
	"fmt"
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

func impliedEPSGrowthBand(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) (float64, float64, bool) {
	peg := pegRatioValue(q, f)
	if peg <= 0 {
		return 0, 0, false
	}

	values := make([]float64, 0, 2)
	if f.ForwardPE > 0 {
		values = append(values, f.ForwardPE/peg/100)
	}
	if f.TrailingPE > 0 {
		values = append(values, f.TrailingPE/peg/100)
	}
	if len(values) == 0 {
		return 0, 0, false
	}
	low, high := values[0], values[0]
	for _, v := range values[1:] {
		if v < low {
			low = v
		}
		if v > high {
			high = v
		}
	}
	return low, high, true
}

func impliedEPSGrowthBandText(q domain.QuoteSnapshot, f domain.FundamentalsSnapshot) string {
	low, high, ok := impliedEPSGrowthBand(q, f)
	if !ok {
		return "-"
	}
	if low == high {
		return formatImpliedGrowthPercent(low * 100)
	}
	return formatImpliedGrowthPercent(low*100) + "-" + formatImpliedGrowthPercent(high*100)
}

func formatImpliedGrowthPercent(v float64) string {
	return fmt.Sprintf("%.0f%%", v)
}
