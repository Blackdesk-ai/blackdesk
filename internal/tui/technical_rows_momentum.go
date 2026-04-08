package tui

func technicalMomentumRows(s technicalSnapshot) []marketTableRow {
	rsiMove := 0.0
	switch {
	case s.rsi14 > 70:
		rsiMove = 1
	case s.rsi14 < 30 && s.rsi14 > 0:
		rsiMove = -1
	}
	return []marketTableRow{
		{name: "RSI 14", price: formatMetricFloat(s.rsi14), chg: rsiLabel(s.rsi14), move: rsiMove, styled: rsiMove != 0},
		{name: "1M return", price: formatSignedPercentRatio(s.ret1M), chg: "", move: s.ret1M, styled: s.ret1M != 0},
		{name: "3M return", price: formatSignedPercentRatio(s.ret3M), chg: "", move: s.ret3M, styled: s.ret3M != 0},
		{name: "12M return", price: formatSignedPercentRatio(s.ret12M), chg: "", move: s.ret12M, styled: s.ret12M != 0},
		{name: "MACD", price: formatMetricSigned(s.macd), chg: macdLabel(s.macd, s.macdSignal), move: s.macdHist, styled: s.macd != 0 || s.macdSignal != 0},
		{name: "Histogram", price: formatMetricSigned(s.macdHist), chg: "", move: s.macdHist, styled: s.macdHist != 0},
	}
}
