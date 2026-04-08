package tui

func regimeSummary(m Model) string {
	vix, vixOK := m.lookupQuote("^VIX")
	spy, spyOK := m.lookupQuote("SPY")
	hyg, hygOK := m.lookupQuote("HYG")
	if !vixOK || !spyOK {
		return "no data"
	}
	vixLevel := vix.Price
	_, spyChg := marketDisplayQuoteLine(spy)
	hygChg := 0.0
	if hygOK {
		_, hygChg = marketDisplayQuoteLine(hyg)
	}

	spyTrendUp := true
	if series, ok := m.marketOpinionHistory["SPY"]; ok && len(series.Candles) >= 200 {
		closes := extractCloses(series.Candles)
		sma := simpleMovingAverage(closes, 200)
		if sma > 0 {
			spyTrendUp = closes[len(closes)-1] >= sma
		}
	}

	switch {
	case vixLevel >= 30 && spyChg < -1:
		return "high stress"
	case vixLevel >= 30:
		return "volatile"
	case vixLevel >= 25:
		if spyChg < 0 {
			return "risk off"
		}
		return "elevated vol"
	case vixLevel >= 20:
		if spyChg >= 0 && hygOK && hygChg >= 0 {
			return "cautious bid"
		}
		if spyChg < -0.5 {
			return "risk off"
		}
		return "elevated vol"
	case vixLevel >= 16:
		if !spyTrendUp {
			if spyChg >= 0 {
				return "counter rally"
			}
			return "trend weak"
		}
		if spyChg >= 0 && (!hygOK || hygChg >= -0.1) {
			return "risk on"
		}
		return "mixed"
	default:
		if !spyTrendUp {
			if spyChg >= 0 {
				return "low vol bounce"
			}
			return "complacent"
		}
		if spyChg >= 0 {
			return "low vol bull"
		}
		return "risk on"
	}
}
