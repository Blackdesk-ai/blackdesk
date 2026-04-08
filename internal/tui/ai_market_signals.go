package tui

import (
	"fmt"
	"strings"

	"blackdesk/internal/domain"
)

func marketRegimeSignals(quotes map[string]domain.QuoteSnapshot, histories map[string]domain.PriceSeries) map[string]string {
	signals := make(map[string]string)
	if q, ok := quotes["SPY"]; ok {
		signals["spy_change_1d"] = fmt.Sprintf("%+.2f%%", q.ChangePercent)
	}
	if q, ok := quotes["HYG"]; ok {
		signals["hyg_change_1d"] = fmt.Sprintf("%+.2f%%", q.ChangePercent)
	}
	if q, ok := quotes["^VIX"]; ok {
		signals["vix_level"] = fmt.Sprintf("%.2f", q.Price)
		signals["vix_change_1d"] = fmt.Sprintf("%+.2f%%", q.ChangePercent)
	}
	if q, ok := quotes["GC=F"]; ok {
		signals["gold_change_1d"] = fmt.Sprintf("%+.2f%%", q.ChangePercent)
	}
	if q, ok := quotes["DX-Y.NYB"]; ok {
		signals["dxy_change_1d"] = fmt.Sprintf("%+.2f%%", q.ChangePercent)
	}
	for _, symbol := range []string{"SPY", "HYG", "^TNX", "2YY=F", "GC=F", "DX-Y.NYB"} {
		series, ok := histories[symbol]
		if !ok {
			continue
		}
		snap := buildTechnicalSnapshot(domain.QuoteSnapshot{Symbol: symbol}, series)
		closes := extractCloses(series.Candles)
		key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(symbol, "^", ""), "=", ""))
		if snap.sma200 > 0 {
			signals[key+"_vs_sma200"] = distanceLabel(snap.price, snap.sma200)
		}
		if ret := lookbackReturn(closes, 5); ret != 0 {
			signals[key+"_change_1w"] = formatSignedPercentRatio(ret)
		}
		if ret := lookbackReturn(closes, 21); ret != 0 {
			signals[key+"_change_1m"] = formatSignedPercentRatio(ret)
		}
	}
	return signals
}
