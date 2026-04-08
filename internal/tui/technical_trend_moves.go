package tui

func relativeMove(price, level float64) float64 {
	if price == 0 || level == 0 {
		return 0
	}
	return (price - level) / level
}

func trendStackMove(s technicalSnapshot) float64 {
	label := trendStackLabel(s)
	switch label {
	case "bullish":
		return 1
	case "bearish":
		return -1
	default:
		return 0
	}
}

func adxMove(v float64) float64 {
	switch {
	case v == 0:
		return 0
	case v >= 25:
		return 1
	case v >= 20:
		return 0
	default:
		return -1
	}
}

func atrPercentOfPrice(atr, price float64) float64 {
	if atr == 0 || price == 0 {
		return 0
	}
	return atr / price
}

func returnOverVol(ret, hvAnnual, scale float64) float64 {
	if hvAnnual == 0 || scale == 0 {
		return 0
	}
	return ret / (hvAnnual / scale)
}

func compressionMove(v float64) float64 {
	switch {
	case v == 0:
		return 0
	case v <= 0.015:
		return 1
	case v <= 0.03:
		return 0
	default:
		return -1
	}
}
