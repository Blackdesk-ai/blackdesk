package tui

import "math"

func historicalVolatility(values []float64, period int) float64 {
	if period <= 1 || len(values) <= period {
		return 0
	}
	returns := make([]float64, 0, period)
	start := len(values) - period
	if start < 1 {
		start = 1
	}
	for i := start; i < len(values); i++ {
		prev := values[i-1]
		curr := values[i]
		if prev <= 0 || curr <= 0 {
			continue
		}
		returns = append(returns, math.Log(curr/prev))
	}
	if len(returns) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range returns {
		mean += v
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, v := range returns {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(returns) - 1)
	return math.Sqrt(variance) * math.Sqrt(252)
}

func historicalVolatilityRank(values []float64, hvPeriod, lookback int) (float64, bool) {
	if hvPeriod <= 1 || lookback <= 1 || len(values) < hvPeriod+lookback {
		return 0, false
	}
	current := historicalVolatility(values, hvPeriod)
	if current == 0 {
		return 0, false
	}
	start := len(values) - lookback
	if start < hvPeriod {
		start = hvPeriod
	}
	low := 0.0
	high := 0.0
	haveRange := false
	for end := start; end <= len(values); end++ {
		hv := historicalVolatility(values[:end], hvPeriod)
		if hv == 0 {
			continue
		}
		if !haveRange || hv < low {
			low = hv
		}
		if !haveRange || hv > high {
			high = hv
		}
		haveRange = true
	}
	if !haveRange || high <= low {
		return 0, false
	}
	rank := ((current - low) / (high - low)) * 100
	return clampFloat(rank, 0, 100), true
}

func historicalVolatilityPercentile(values []float64, hvPeriod, lookback int) (float64, bool) {
	if hvPeriod <= 1 || lookback <= 1 || len(values) < hvPeriod+lookback {
		return 0, false
	}
	current := historicalVolatility(values, hvPeriod)
	if current == 0 {
		return 0, false
	}
	start := len(values) - lookback
	if start < hvPeriod {
		start = hvPeriod
	}
	below := 0
	total := 0
	for end := start; end <= len(values); end++ {
		hv := historicalVolatility(values[:end], hvPeriod)
		if hv == 0 {
			continue
		}
		total++
		if hv < current {
			below++
		}
	}
	if total == 0 {
		return 0, false
	}
	return (float64(below) / float64(total)) * 100, true
}

func rollingReturnZScore(values []float64, window, history int) (zScore, tailProb float64, samples int) {
	if window <= 0 || history <= window || len(values) <= history+window {
		return 0, 0, 0
	}
	current := lookbackReturn(values, window)
	start := len(values) - (history + window)
	if start < 0 {
		start = 0
	}
	series := make([]float64, 0, history)
	for i := start + window; i < len(values); i++ {
		base := values[i-window]
		last := values[i]
		if base == 0 {
			continue
		}
		series = append(series, (last-base)/base)
	}
	if len(series) < 2 {
		return 0, 0, len(series)
	}
	mean := 0.0
	for _, v := range series {
		mean += v
	}
	mean /= float64(len(series))
	variance := 0.0
	for _, v := range series {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(series) - 1)
	stdDev := math.Sqrt(variance)
	if stdDev == 0 {
		return 0, 0, len(series)
	}
	zScore = (current - mean) / stdDev
	tailProb = math.Erfc(math.Abs(zScore) / math.Sqrt2)
	return zScore, tailProb, len(series)
}
