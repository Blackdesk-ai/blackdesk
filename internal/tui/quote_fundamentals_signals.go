package tui

import (
	"blackdesk/internal/domain"
	"blackdesk/internal/ui"
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
