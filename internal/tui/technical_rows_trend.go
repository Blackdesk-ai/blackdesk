package tui

func technicalTrendRows(s technicalSnapshot) []marketTableRow {
	trendMove := trendStackMove(s)
	return []marketTableRow{
		{name: "ADX 14", price: formatMetricFloat(s.adx14), chg: adxLabel(s.adx14), move: adxMove(s.adx14), styled: s.adx14 > 0},
		{name: "EMA 12", price: formatMoneyDash(s.ema12), chg: distanceLabel(s.price, s.ema12), move: relativeMove(s.price, s.ema12), styled: s.price > 0 && s.ema12 > 0},
		{name: "EMA 26", price: formatMoneyDash(s.ema26), chg: distanceLabel(s.price, s.ema26), move: relativeMove(s.price, s.ema26), styled: s.price > 0 && s.ema26 > 0},
		{name: "SMA 20", price: formatMoneyDash(s.sma20), chg: distanceLabel(s.price, s.sma20), move: relativeMove(s.price, s.sma20), styled: s.price > 0 && s.sma20 > 0},
		{name: "SMA 50", price: formatMoneyDash(s.sma50), chg: distanceLabel(s.price, s.sma50), move: relativeMove(s.price, s.sma50), styled: s.price > 0 && s.sma50 > 0},
		{name: "SMA 200", price: formatMoneyDash(s.sma200), chg: distanceLabel(s.price, s.sma200), move: relativeMove(s.price, s.sma200), styled: s.price > 0 && s.sma200 > 0},
		{name: "Trend stack", price: trendStackLabel(s), chg: moveArrow(trendMove), move: trendMove, styled: true},
	}
}
