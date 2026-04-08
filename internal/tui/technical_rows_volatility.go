package tui

import "blackdesk/internal/ui"

func technicalVolatilityRows(s technicalSnapshot) []marketTableRow {
	atrPct := atrPercentOfPrice(s.atr14, s.price)
	return []marketTableRow{
		{name: "ATR 14", price: formatMoneyDash(s.atr14), chg: compressionLabel(atrPct), move: compressionMove(atrPct), styled: s.atr14 > 0 && s.price > 0},
		{name: "HV Rank", price: formatOptionalPercent(s.hvRank21/100, s.hvRankReady), chg: hvPercentileLabel(s.hvPct21, s.hvPctReady), move: hvPercentileMove(s.hvPct21, s.hvPctReady), styled: s.hvRankReady && s.hvPctReady},
		{name: "HV Pctl", price: formatOptionalPercent(s.hvPct21/100, s.hvPctReady), chg: hvPercentileLabel(s.hvPct21, s.hvPctReady), move: hvPercentileMove(s.hvPct21, s.hvPctReady), styled: s.hvPctReady},
		{name: "HV 21", price: percentDash(s.hv21), chg: hvLabel(s.hv21, s.hv63, s.hv252), move: hvMove(s.hv21, s.hv63, s.hv252), styled: s.hv21 > 0},
		{name: "HV 63", price: percentDash(s.hv63), chg: "baseline", move: 0, styled: false},
		{name: "HV 252", price: percentDash(s.hv252), chg: "baseline", move: 0, styled: false},
	}
}

func technicalVolumeRows(s technicalSnapshot) []marketTableRow {
	volMove := 0.0
	if s.volumeRatio20 > 0 {
		volMove = s.volumeRatio20 - 1
	}
	return []marketTableRow{
		{name: "Last volume", price: ui.FormatCompactInt(s.lastVolume), chg: "", move: 0, styled: false},
		{name: "RVOL", price: formatRatio(s.volumeRatio20), chg: volumeLabel(s.volumeRatio20), move: volMove, styled: s.volumeRatio20 > 0},
	}
}
