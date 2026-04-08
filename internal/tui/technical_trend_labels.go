package tui

func breakoutLabel(price, level float64, resistance bool) string {
	if price == 0 || level == 0 {
		return "-"
	}
	if resistance {
		if price > level {
			return "breakout"
		}
		return "below high"
	}
	if price < level {
		return "breakdown"
	}
	return "above low"
}

func rsiLabel(v float64) string {
	switch {
	case v == 0:
		return "-"
	case v >= 70:
		return "overbought"
	case v <= 30:
		return "oversold"
	default:
		return "neutral"
	}
}

func macdLabel(line, signal float64) string {
	if line == 0 && signal == 0 {
		return "-"
	}
	if line > signal {
		return "bullish"
	}
	if line < signal {
		return "bearish"
	}
	return "flat"
}

func trendStackLabel(s technicalSnapshot) string {
	if s.sma20 == 0 || s.sma50 == 0 || s.sma200 == 0 {
		return "-"
	}
	switch {
	case s.price > s.sma20 && s.sma20 > s.sma50 && s.sma50 > s.sma200:
		return "bullish"
	case s.price < s.sma20 && s.sma20 < s.sma50 && s.sma50 < s.sma200:
		return "bearish"
	default:
		return "mixed"
	}
}

func adxLabel(v float64) string {
	switch {
	case v == 0:
		return "-"
	case v >= 40:
		return "very strong"
	case v >= 25:
		return "strong"
	case v >= 20:
		return "building"
	default:
		return "weak"
	}
}

func rangePositionLabel(v float64) string {
	switch {
	case v == 0:
		return "-"
	case v >= 0.8:
		return "near highs"
	case v <= 0.2:
		return "near lows"
	default:
		return "mid range"
	}
}

func historyDepthLabel(count int) string {
	switch {
	case count >= 200:
		return "full"
	case count >= 60:
		return "usable"
	case count > 0:
		return "light"
	default:
		return "-"
	}
}

func volumeLabel(v float64) string {
	switch {
	case v == 0:
		return "-"
	case v >= 1.5:
		return "heavy"
	case v >= 1.1:
		return "active"
	case v < 0.8:
		return "light"
	default:
		return "normal"
	}
}

func compressionLabel(v float64) string {
	switch {
	case v == 0:
		return "-"
	case v <= 0.015:
		return "tight"
	case v <= 0.03:
		return "normal"
	default:
		return "expanded"
	}
}
